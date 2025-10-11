package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"markdown-note-taking-app/internal/models"
)

// noteRepository implements NoteRepository
type noteRepository struct {
	db *DB
}

// NewNoteRepository creates a new note repository
func NewNoteRepository(db *DB) NoteRepository {
	return &noteRepository{db: db}
}

// Create inserts a new note into the database
func (r *noteRepository) Create(note *models.Note) error {
	query := `
		INSERT INTO notes (title, content, created_at, updated_at)
		VALUES (?, ?, ?, ?)`

	result, err := r.db.Exec(query, note.Title, note.Content, note.CreatedAt, note.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted note ID: %w", err)
	}

	note.ID = int(id)
	return nil
}

// GetByID retrieves a note by its ID
func (r *noteRepository) GetByID(id int) (*models.Note, error) {
	query := `
		SELECT id, title, content, created_at, updated_at
		FROM notes
		WHERE id = ?`

	note := &models.Note{}
	var createdAt, updatedAt string

	err := r.db.QueryRow(query, id).Scan(
		&note.ID, &note.Title, &note.Content, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Parse timestamps
	note.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	note.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	// Load tags
	tags, err := r.getNoteTags(note.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	note.Tags = tags

	return note, nil
}

// GetAll retrieves all notes with optional filtering
func (r *noteRepository) GetAll(filter models.NoteFilter) ([]*models.Note, error) {
	query := `
		SELECT DISTINCT n.id, n.title, n.content, n.created_at, n.updated_at
		FROM notes n`

	args := []any{}
	conditions := []string{}

	// Add search condition
	if filter.SearchQuery != "" {
		conditions = append(conditions, "(n.title LIKE ? OR n.content LIKE ?)")
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Add tag filter
	if len(filter.TagIDs) > 0 {
		placeholders := strings.Repeat("?,", len(filter.TagIDs))
		placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
		conditions = append(conditions, fmt.Sprintf("n.id IN (SELECT note_id FROM note_tags WHERE tag_id IN (%s))", placeholders))
		for _, tagID := range filter.TagIDs {
			args = append(args, tagID)
		}
	}

	// Add WHERE clause if we have conditions
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	query += " ORDER BY n.updated_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note := &models.Note{}
		var createdAt, updatedAt string

		err := rows.Scan(&note.ID, &note.Title, &note.Content, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		// Parse timestamps
		note.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		note.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		// Load tags for this note
		tags, err := r.getNoteTags(note.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load tags for note %d: %w", note.ID, err)
		}
		note.Tags = tags

		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// Update modifies an existing note
func (r *noteRepository) Update(note *models.Note) error {
	query := `
		UPDATE notes
		SET title = ?, content = ?, updated_at = ?
		WHERE id = ?`

	note.UpdatedAt = time.Now()
	result, err := r.db.Exec(query, note.Title, note.Content, note.UpdatedAt, note.ID)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note with ID %d not found", note.ID)
	}

	return nil
}

// Delete removes a note from the database
func (r *noteRepository) Delete(id int) error {
	query := `DELETE FROM notes WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note with ID %d not found", id)
	}

	return nil
}

// Search performs a full-text search on notes
func (r *noteRepository) Search(query string, limit int) ([]*models.Note, error) {
	filter := models.NoteFilter{
		SearchQuery: query,
		Limit:       limit,
	}
	return r.GetAll(filter)
}

// GetByTag retrieves all notes with a specific tag
func (r *noteRepository) GetByTag(tagID int) ([]*models.Note, error) {
	filter := models.NoteFilter{
		TagIDs: []int{tagID},
	}
	return r.GetAll(filter)
}

// AddTag associates a tag with a note
func (r *noteRepository) AddTag(noteID, tagID int) error {
	query := `
		INSERT OR IGNORE INTO note_tags (note_id, tag_id)
		VALUES (?, ?)`

	_, err := r.db.Exec(query, noteID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to note: %w", err)
	}

	return nil
}

// RemoveTag removes a tag association from a note
func (r *noteRepository) RemoveTag(noteID, tagID int) error {
	query := `DELETE FROM note_tags WHERE note_id = ? AND tag_id = ?`

	result, err := r.db.Exec(query, noteID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag association not found for note %d and tag %d", noteID, tagID)
	}

	return nil
}

// getNoteTags retrieves all tags for a specific note
func (r *noteRepository) getNoteTags(noteID int) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name
		FROM tags t
		JOIN note_tags nt ON t.id = nt.tag_id
		WHERE nt.note_id = ?
		ORDER BY t.name`

	rows, err := r.db.Query(query, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query note tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

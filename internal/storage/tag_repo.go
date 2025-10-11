package storage

import (
	"database/sql"
	"fmt"

	"markdown-note-taking-app/internal/models"
)

// tagRepository implements TagRepository
type tagRepository struct {
	db *DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *DB) TagRepository {
	return &tagRepository{db: db}
}

// Create inserts a new tag into the database
func (r *tagRepository) Create(name string) (*models.Tag, error) {
	query := `INSERT INTO tags (name) VALUES (?)`

	result, err := r.db.Exec(query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get inserted tag ID: %w", err)
	}

	return &models.Tag{ID: int(id), Name: name}, nil
}

// GetByID retrieves a tag by its ID
func (r *tagRepository) GetByID(id int) (*models.Tag, error) {
	query := `SELECT id, name FROM tags WHERE id = ?`

	tag := &models.Tag{}
	err := r.db.QueryRow(query, id).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

// GetAll retrieves all tags
func (r *tagRepository) GetAll() ([]*models.Tag, error) {
	query := `SELECT id, name FROM tags ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		tag := &models.Tag{}
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// GetByName retrieves a tag by its name
func (r *tagRepository) GetByName(name string) (*models.Tag, error) {
	query := `SELECT id, name FROM tags WHERE name = ?`

	tag := &models.Tag{}
	err := r.db.QueryRow(query, name).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag with name '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

// Update modifies an existing tag
func (r *tagRepository) Update(tag *models.Tag) error {
	query := `UPDATE tags SET name = ? WHERE id = ?`

	result, err := r.db.Exec(query, tag.Name, tag.ID)
	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag with ID %d not found", tag.ID)
	}

	return nil
}

// Delete removes a tag from the database
func (r *tagRepository) Delete(id int) error {
	query := `DELETE FROM tags WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag with ID %d not found", id)
	}

	return nil
}

// GetNoteTags retrieves all tags for a specific note
func (r *tagRepository) GetNoteTags(noteID int) ([]*models.Tag, error) {
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

	var tags []*models.Tag
	for rows.Next() {
		tag := &models.Tag{}
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

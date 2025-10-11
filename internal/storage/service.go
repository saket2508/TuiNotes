package storage

import (
	"fmt"

	"markdown-note-taking-app/internal/models"
)

// Service provides high-level operations combining repositories
type Service struct {
	db    *DB
	notes NoteRepository
	tags  TagRepository
}

// NewService creates a new storage service
func NewService(dbPath string) (*Service, error) {
	db, err := NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return &Service{
		db:    db,
		notes: NewNoteRepository(db),
		tags:  NewTagRepository(db),
	}, nil
}

// Close closes the database connection
func (s *Service) Close() error {
	return s.db.Close()
}

// Note operations

// CreateNote creates a new note
func (s *Service) CreateNote(title, content string) (*models.Note, error) {
	note := models.NewNote(title, content)
	if err := s.notes.Create(note); err != nil {
		return nil, err
	}
	return note, nil
}

// GetNote retrieves a note by ID
func (s *Service) GetNote(id int) (*models.Note, error) {
	return s.notes.GetByID(id)
}

// GetAllNotes retrieves all notes with optional filtering
func (s *Service) GetAllNotes(filter models.NoteFilter) ([]*models.Note, error) {
	return s.notes.GetAll(filter)
}

// UpdateNote updates an existing note
func (s *Service) UpdateNote(note *models.Note) error {
	return s.notes.Update(note)
}

// DeleteNote deletes a note
func (s *Service) DeleteNote(id int) error {
	return s.notes.Delete(id)
}

// SearchNotes performs a search on notes
func (s *Service) SearchNotes(query string, limit int) ([]*models.Note, error) {
	return s.notes.Search(query, limit)
}

// Tag operations

// CreateTag creates a new tag
func (s *Service) CreateTag(name string) (*models.Tag, error) {
	return s.tags.Create(name)
}

// GetTag retrieves a tag by ID
func (s *Service) GetTag(id int) (*models.Tag, error) {
	return s.tags.GetByID(id)
}

// GetAllTags retrieves all tags
func (s *Service) GetAllTags() ([]*models.Tag, error) {
	return s.tags.GetAll()
}

// GetOrCreateTag gets a tag by name or creates it if it doesn't exist
func (s *Service) GetOrCreateTag(name string) (*models.Tag, error) {
	tag, err := s.tags.GetByName(name)
	if err != nil {
		// Tag doesn't exist, create it
		tag, err = s.tags.Create(name)
		if err != nil {
			return nil, err
		}
	}
	return tag, nil
}

// UpdateTag updates an existing tag
func (s *Service) UpdateTag(tag *models.Tag) error {
	return s.tags.Update(tag)
}

// DeleteTag deletes a tag
func (s *Service) DeleteTag(id int) error {
	return s.tags.Delete(id)
}

// Note-Tag operations

// AddTagToNote adds a tag to a note
func (s *Service) AddTagToNote(noteID int, tagName string) error {
	tag, err := s.GetOrCreateTag(tagName)
	if err != nil {
		return err
	}
	return s.notes.AddTag(noteID, tag.ID)
}

// RemoveTagFromNote removes a tag from a note
func (s *Service) RemoveTagFromNote(noteID, tagID int) error {
	return s.notes.RemoveTag(noteID, tagID)
}

// GetNotesByTag retrieves all notes with a specific tag
func (s *Service) GetNotesByTag(tagID int) ([]*models.Note, error) {
	return s.notes.GetByTag(tagID)
}

// GetNoteTags retrieves all tags for a specific note
func (s *Service) GetNoteTags(noteID int) ([]*models.Tag, error) {
	return s.tags.GetNoteTags(noteID)
}

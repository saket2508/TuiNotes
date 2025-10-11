package storage

import (
	"markdown-note-taking-app/internal/models"
)

// NoteRepository defines the interface for note operations
type NoteRepository interface {
	Create(note *models.Note) error
	GetByID(id int) (*models.Note, error)
	GetAll(filter models.NoteFilter) ([]*models.Note, error)
	Update(note *models.Note) error
	Delete(id int) error
	Search(query string, limit int) ([]*models.Note, error)
	GetByTag(tagID int) ([]*models.Note, error)
	AddTag(noteID, tagID int) error
	RemoveTag(noteID, tagID int) error
}

// TagRepository defines the interface for tag operations
type TagRepository interface {
	Create(name string) (*models.Tag, error)
	GetByID(id int) (*models.Tag, error)
	GetAll() ([]*models.Tag, error)
	GetByName(name string) (*models.Tag, error)
	Update(tag *models.Tag) error
	Delete(id int) error
	GetNoteTags(noteID int) ([]*models.Tag, error)
}

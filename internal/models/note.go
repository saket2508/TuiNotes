package models

import (
	"time"
)

// Note represents a markdown note
type Note struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Tags      []Tag     `json:"tags,omitempty" db:"-"`
}

// Tag represents a tag that can be assigned to notes
type Tag struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// NoteFilter represents filters for querying notes
type NoteFilter struct {
	SearchQuery string
	TagIDs      []int
	Limit       int
	Offset      int
}

// NewNote creates a new note with timestamps
func NewNote(title, content string) *Note {
	now := time.Now()
	return &Note{
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      []Tag{},
	}
}

// UpdateContent updates the note content and timestamp
func (n *Note) UpdateContent(content string) {
	n.Content = content
	n.UpdatedAt = time.Now()
}

// UpdateTitle updates the note title and timestamp
func (n *Note) UpdateTitle(title string) {
	n.Title = title
	n.UpdatedAt = time.Now()
}

// AddTag adds a tag to the note
func (n *Note) AddTag(tag Tag) {
	// Check if tag already exists
	for _, existingTag := range n.Tags {
		if existingTag.ID == tag.ID {
			return
		}
	}
	n.Tags = append(n.Tags, tag)
}

// RemoveTag removes a tag from the note
func (n *Note) RemoveTag(tagID int) {
	for i, tag := range n.Tags {
		if tag.ID == tagID {
			n.Tags = append(n.Tags[:i], n.Tags[i+1:]...)
			break
		}
	}
}

// HasTag checks if note has a specific tag
func (n *Note) HasTag(tagID int) bool {
	for _, tag := range n.Tags {
		if tag.ID == tagID {
			return true
		}
	}
	return false
}

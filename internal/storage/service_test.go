package storage

import (
	"os"
	"testing"
)

func TestService(t *testing.T) {
	// Create a temporary database for testing
	tmpFile, err := os.CreateTemp("", "notes_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create service
	service, err := NewService(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	// Test creating a note
	note, err := service.CreateNote("Test Note", "# Hello World\n\nThis is a test note.")
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	if note.ID == 0 {
		t.Error("Expected note ID to be set")
	}

	if note.Title != "Test Note" {
		t.Errorf("Expected title 'Test Note', got '%s'", note.Title)
	}

	// Test retrieving the note
	retrievedNote, err := service.GetNote(note.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve note: %v", err)
	}

	if retrievedNote.Title != note.Title {
		t.Errorf("Expected title '%s', got '%s'", note.Title, retrievedNote.Title)
	}

	if retrievedNote.Content != note.Content {
		t.Errorf("Expected content '%s', got '%s'", note.Content, retrievedNote.Content)
	}

	// Test creating a tag
	tag, err := service.CreateTag("test")
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	if tag.Name != "test" {
		t.Errorf("Expected tag name 'test', got '%s'", tag.Name)
	}

	// Test adding tag to note
	err = service.AddTagToNote(note.ID, "test")
	if err != nil {
		t.Fatalf("Failed to add tag to note: %v", err)
	}

	// Test retrieving note with tags
	noteWithTags, err := service.GetNote(note.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve note with tags: %v", err)
	}

	if len(noteWithTags.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(noteWithTags.Tags))
	}

	if noteWithTags.Tags[0].Name != "test" {
		t.Errorf("Expected tag name 'test', got '%s'", noteWithTags.Tags[0].Name)
	}

	// Test searching notes
	results, err := service.SearchNotes("Hello", 10)
	if err != nil {
		t.Fatalf("Failed to search notes: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(results))
	}

	t.Logf("Storage layer test passed! Created note ID: %d, Tag ID: %d", note.ID, tag.ID)
}

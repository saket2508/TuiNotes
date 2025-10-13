package ui

import (
	"strings"
	"testing"
)

func TestNoteEditorView(t *testing.T) {
	app := &App{} // Mock app for testing
	editor := NewNoteEditorModel(app)

	// Set some test data
	editor.title = "Test Title"
	editor.content = "Test content\nwith multiple lines"
	editor.focused = 0 // Title should be active
	editor.width = 80
	editor.height = 24

	view := editor.View()

	// Check that the view contains expected elements
	if !strings.Contains(view, "Create Note") {
		t.Error("View should contain 'Create Note' title")
	}

	if !strings.Contains(view, "[*] Title:") {
		t.Error("View should show title as active with [*]")
	}

	if !strings.Contains(view, "[ ] Content:") {
		t.Error("View should show content as inactive with [ ]")
	}

	if !strings.Contains(view, "Test Title") {
		t.Error("View should contain the test title")
	}

	if !strings.Contains(view, "Test content") {
		t.Error("View should contain the test content")
	}

	t.Logf("Editor view output:\n%s", view)
}

func TestNoteEditorFocusSwitching(t *testing.T) {
	app := &App{}
	editor := NewNoteEditorModel(app)
	editor.width = 80
	editor.height = 24

	// Test title focus (focused = false)
	editor.focused = 0
	view := editor.View()

	if !strings.Contains(view, "[*] Title:") {
		t.Error("Title should be marked as active when focused = false")
	}

	if !strings.Contains(view, "[ ] Content:") {
		t.Error("Content should be marked as inactive when focused = false")
	}

	// Test content focus (focused = true)
	editor.focused = 0
	view = editor.View()

	if !strings.Contains(view, "[ ] Title:") {
		t.Error("Title should be marked as inactive when focused = true")
	}

	if !strings.Contains(view, "[*] Content:") {
		t.Error("Content should be marked as active when focused = true")
	}
}

package app

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewComment(t *testing.T) {
	t.Run("Create comment without parent", func(t *testing.T) {
		text := "Hello, world!"
		comment, err := NewComment("", text)
		assert.NoError(t, err)
		assert.NotNil(t, comment)
		assert.Equal(t, text, comment.Text)
		assert.Nil(t, comment.ParentID)
		assert.NotEqual(t, uuid.Nil, comment.ID)
	})

	t.Run("Create comment with valid parent", func(t *testing.T) {
		parentID := uuid.New().String()
		text := "Reply to comment"
		comment, err := NewComment(parentID, text)
		assert.NoError(t, err)
		assert.NotNil(t, comment)
		assert.Equal(t, text, comment.Text)
		assert.NotNil(t, comment.ParentID)
		assert.Equal(t, parentID, comment.ParentID.String())
		assert.NotEqual(t, uuid.Nil, comment.ID)
	})

	t.Run("Fail on empty text", func(t *testing.T) {
		comment, err := NewComment("", "")
		assert.Error(t, err)
		assert.Nil(t, comment)
	})

	t.Run("Fail on invalid parent UUID", func(t *testing.T) {
		comment, err := NewComment("invalid-uuid", "Text")
		assert.Error(t, err)
		assert.Nil(t, comment)
	})
}

func TestBuildTree(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()
	comments := []Comment{
		{ID: rootID, Text: "Root comment"},
		{ID: childID, Text: "Child comment", ParentID: &rootID},
		{ID: uuid.New(), Text: "Another root comment"},
	}

	tree := BuildTree(comments, nil)
	assert.Len(t, tree, 2) // два корневых комментария

	// Проверка первого корня
	assert.Equal(t, "Root comment", tree[0].Text)
	assert.Len(t, tree[0].Children, 1)
	assert.Equal(t, "Child comment", tree[0].Children[0].Text)

	// Проверка второго корня
	assert.Equal(t, "Another root comment", tree[1].Text)
	assert.Len(t, tree[1].Children, 0)
}

func TestFilterTreeByText(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()
	grandChildID := uuid.New()

	comments := []Comment{
		{ID: rootID, Text: "Root comment"},
		{ID: childID, Text: "Child comment", ParentID: &rootID},
		{ID: grandChildID, Text: "Grandchild filter me", ParentID: &childID},
		{ID: uuid.New(), Text: "Another root comment"},
	}

	tree := BuildTree(comments, nil)

	// Фильтрация по слову "filter"
	filtered := FilterTreeByText(tree, "filter")

	// Должен вернуть только первый корень с дочерним узлом, который содержит "filter"
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Root comment", filtered[0].Text)
	assert.Len(t, filtered[0].Children, 1)
	assert.Equal(t, "Child comment", filtered[0].Children[0].Text)
	assert.Len(t, filtered[0].Children[0].Children, 1)
	assert.Equal(t, "Grandchild filter me", filtered[0].Children[0].Children[0].Text)

	// Фильтрация по слову, которого нет
	filteredNone := FilterTreeByText(tree, "nonexistent")
	assert.Len(t, filteredNone, 0)
}

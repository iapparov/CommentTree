package app

import (
	domain "commentTree/internal/app/domain"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockDb struct {
	mock.Mock
}

func (m *MockDb) SaveComment(text, parentID string) (*domain.Comment, error) {
	args := m.Called(text, parentID)
	return args.Get(0).(*domain.Comment), args.Error(1)
}

func (m *MockDb) GetComments(parentId string, sortAsc string, page, pageSize int) ([]domain.Comment, error) {
	args := m.Called(parentId, sortAsc, page, pageSize)
	return args.Get(0).([]domain.Comment), args.Error(1)
}

func (m *MockDb) SearchComments(text string, sortAsc string, page, pageSize int) ([]domain.Comment, error) {
	args := m.Called(text, sortAsc, page, pageSize)
	return args.Get(0).([]domain.Comment), args.Error(1)
}

func (m *MockDb) DeleteComments(parentId string) error {
	args := m.Called(parentId)
	return args.Error(0)
}

func TestCommentService_CreateComment(t *testing.T) {
	mockDb := new(MockDb)
	service := NewCommentService(mockDb)

	comment := &domain.Comment{ID: uuid.New(), Text: "Test comment"}

	mockDb.On("SaveComment", "Test comment", "").Return(comment, nil)

	result, err := service.CreateComment("Test comment", "")
	assert.NoError(t, err)
	assert.Equal(t, comment, result)
	mockDb.AssertExpectations(t)
}

func TestCommentService_GetComments(t *testing.T) {
	mockDb := new(MockDb)
	service := NewCommentService(mockDb)

	rootID := uuid.New()
	comments := []domain.Comment{
		{ID: rootID, Text: "Root comment"},
	}

	mockDb.On("GetComments", rootID.String(), "asc", 1, 10).Return(comments, nil)

	result, err := service.GetComments(rootID.String(), "asc", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Root comment", result[0].Text)
	mockDb.AssertExpectations(t)
}

func TestCommentService_SearchComments(t *testing.T) {
	mockDb := new(MockDb)
	service := NewCommentService(mockDb)

	comments := []domain.Comment{
		{ID: uuid.New(), Text: "Hello world"},
	}

	mockDb.On("SearchComments", "hello", "asc", 1, 10).Return(comments, nil)

	result, err := service.SearchComments("hello", "", "asc", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Hello world", result[0].Text)
	mockDb.AssertExpectations(t)
}

func TestCommentService_DeleteComments(t *testing.T) {
	mockDb := new(MockDb)
	service := NewCommentService(mockDb)

	id := uuid.New().String()
	mockDb.On("DeleteComments", id).Return(nil)

	err := service.DeleteComments(id)
	assert.NoError(t, err)
	mockDb.AssertExpectations(t)
}

func TestCommentService_DeleteComments_InvalidUUID(t *testing.T) {
	mockDb := new(MockDb)
	service := NewCommentService(mockDb)

	err := service.DeleteComments("invalid-uuid")
	assert.Error(t, err)
}

func TestCommentService_SearchComments_WithParentID(t *testing.T) {
	mockDb := new(MockDb)
	service := NewCommentService(mockDb)

	parentID := uuid.New().String()
	childID := uuid.New()

	// Подготавливаем дерево, которое вернёт GetComments
	rootComment := domain.Comment{ID: uuid.MustParse(parentID), Text: "Root comment"}
	childComment := domain.Comment{ID: uuid.MustParse(childID.String()), Text: "Child filter me", ParentID: &rootComment.ID}

	mockDb.On("GetComments", parentID, "asc", 1, 10).Return([]domain.Comment{rootComment, childComment}, nil)

	result, err := service.SearchComments("filter", parentID, "asc", 1, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1) // должен вернуть корневой узел
	assert.Equal(t, "Root comment", result[0].Text)
	assert.Len(t, result[0].Children, 1)
	assert.Equal(t, "Child filter me", result[0].Children[0].Text)

	mockDb.AssertExpectations(t)
}

func TestCommentService_SearchComments_WithParentID_Error(t *testing.T) {
	mockDb := new(MockDb)
	service := NewCommentService(mockDb)

	parentID := uuid.New().String()

	mockDb.On("GetComments", parentID, "asc", 1, 10).Return([]domain.Comment{}, errors.New("db error"))

	result, err := service.SearchComments("filter", parentID, "asc", 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockDb.AssertExpectations(t)
}

package web

import (
	"bytes"
	"commentTree/internal/app/domain"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockCommentService struct {
	createCommentFunc  func(text, parentID string) (*app.Comment, error)
	getCommentsFunc    func(parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error)
	searchCommentsFunc func(text string, parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error)
	deleteCommentsFunc func(id string) error
}

func (m *MockCommentService) CreateComment(text, parentID string) (*app.Comment, error) {
	return m.createCommentFunc(text, parentID)
}

func (m *MockCommentService) GetComments(parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error) {
	return m.getCommentsFunc(parentId, sortAsc, page, pageSize)
}

func (m *MockCommentService) SearchComments(text string, parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error) {
	return m.searchCommentsFunc(text, parentId, sortAsc, page, pageSize)
}

func (m *MockCommentService) DeleteComments(id string) error {
	return m.deleteCommentsFunc(id)
}

func TestCreateComment_Success(t *testing.T) {
	mock := &MockCommentService{
		createCommentFunc: func(text, parentID string) (*app.Comment, error) {
			id := uuid.New()
			parentUUID := uuid.Nil
			if parentID != "" {
				parentUUID, _ = uuid.Parse(parentID)
			}
			return &app.Comment{ID: id, Text: text, ParentID: &parentUUID}, nil
		},
	}
	handler := NewCommentHandler(mock)

	body := CommentReqCreate{ParentId: "parent1", Text: "Test comment"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.CreateComment(ctx)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestCreateComment_InvalidJSON(t *testing.T) {
	mock := &MockCommentService{}
	handler := NewCommentHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.CreateComment(ctx)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateComment_MissingText(t *testing.T) {
	mock := &MockCommentService{}
	handler := NewCommentHandler(mock)

	body := CommentReqCreate{ParentId: "parent1"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.CreateComment(ctx)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateComment_ServiceError(t *testing.T) {
	mock := &MockCommentService{
		createCommentFunc: func(text, parentID string) (*app.Comment, error) {
			return nil, errors.New("database error")
		},
	}
	handler := NewCommentHandler(mock)

	body := CommentReqCreate{ParentId: "parent1", Text: "Test comment"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.CreateComment(ctx)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestDeleteComments_Success(t *testing.T) {
	mock := &MockCommentService{
		deleteCommentsFunc: func(id string) error {
			return nil
		},
	}
	handler := NewCommentHandler(mock)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Params = []gin.Param{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}

	handler.DeleteComments(ctx)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestDeleteComments_MissingId(t *testing.T) {
	mock := &MockCommentService{}
	handler := NewCommentHandler(mock)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	handler.DeleteComments(ctx)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteComments_ServiceError(t *testing.T) {
	mock := &MockCommentService{
		deleteCommentsFunc: func(id string) error {
			return errors.New("ошибка БД")
		},
	}
	handler := NewCommentHandler(mock)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Params = []gin.Param{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}

	handler.DeleteComments(ctx)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestGetComments_Success(t *testing.T) {
	mock := &MockCommentService{
		getCommentsFunc: func(parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error) {
			return []app.CommentNode{}, nil
		},
	}
	handler := NewCommentHandler(mock)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/comments?parent=123&page=1&page_size=10", nil)

	handler.GetComments(ctx)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetComments_WithSearch(t *testing.T) {
	mock := &MockCommentService{
		searchCommentsFunc: func(text string, parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error) {
			return []app.CommentNode{}, nil
		},
	}
	handler := NewCommentHandler(mock)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/comments?search=test&parent=123&page=1&page_size=10", nil)

	handler.GetComments(ctx)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetComments_ServiceError(t *testing.T) {
	mock := &MockCommentService{
		getCommentsFunc: func(parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error) {
			return nil, errors.New("ошибка БД")
		},
	}
	handler := NewCommentHandler(mock)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/comments?parent=123&page=1&page_size=10", nil)

	handler.GetComments(ctx)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

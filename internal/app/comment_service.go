package app

import (
	"commentTree/internal/app/domain"
	"github.com/google/uuid"
	wbzlog "github.com/wb-go/wbf/zlog"
)

type CommentService struct {
	db DbProvider
}

type DbProvider interface {
	SaveComment(text, parentID string) (*app.Comment, error)
	GetComments(parentId string, sortAsc string, page, pageSize int) ([]app.Comment, error)
	SearchComments(text string, sortAsc string, page, pageSize int) ([]app.Comment, error)
	DeleteComments(parentId string) error
}

func NewCommentService(db DbProvider) *CommentService {
	return &CommentService{
		db: db,
	}
}

func (s *CommentService) CreateComment(text, parentID string) (*app.Comment, error) {
	return s.db.SaveComment(text, parentID)
}

func (s *CommentService) GetComments(parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error) {
	var comments []app.Comment
	var err error

	if parentId == "" {
		comments, err = s.db.GetComments("", sortAsc, page, pageSize)
		if err != nil {
			return nil, err
		}
		nodes := app.BuildTree(comments, nil)
		return nodes, nil
	}

	pID, err := uuid.Parse(parentId)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("invalid parent id")
		return nil, err
	}

	comments, err = s.db.GetComments(parentId, sortAsc, page, pageSize)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to get comments from db")
		return nil, err
	}

	var root *app.Comment
	for _, c := range comments {
		if c.ID == pID {
			root = &c
			break
		}
	}

	if root == nil {
		return nil, nil
	}

	node := app.CommentNode{
		Comment:  *root,
		Children: app.BuildTree(comments, &pID),
	}
	return []app.CommentNode{node}, nil
}

func (s *CommentService) SearchComments(text string, parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error) {
	var nodes []app.CommentNode

	if parentId == "" {
		comments, err := s.db.SearchComments(text, sortAsc, page, pageSize)
		if err != nil {
			return nil, err
		}
		nodes = app.BuildTree(comments, comments[0].ParentID)
	} else {
		tree, err := s.GetComments(parentId, sortAsc, page, pageSize)
		if err != nil {
			return nil, err
		}
		nodes = app.FilterTreeByText(tree, text)
	}
	return nodes, nil
}

func (s *CommentService) DeleteComments(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("invalid id")
		return err
	}
	return s.db.DeleteComments(id)
}

package web

import (
	"commentTree/internal/app/domain"
	wbgin "github.com/wb-go/wbf/ginext"
	"net/http"
	"strconv"
)

type CommentReqCreate struct {
	ParentId string `json:"parent_id"`
	Text     string `json:"text" binding:"required"`
}

type CommentHandler struct {
	commentService CommentService
}

type CommentService interface {
	GetComments(parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error)
	SearchComments(text string, parentId string, sortAsc string, page, pageSize int) ([]app.CommentNode, error)
	DeleteComments(id string) error
	CreateComment(text, parentID string) (*app.Comment, error)
}

func NewCommentHandler(commentService CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// ErrorResponse представляет стандартную ошибку API
type ErrorResponse struct {
	Error string `json:"error" example:"invalid input data"`
}

// CreateComment godoc
// @Summary      Create Comment
// @Description  Создает новый комментарий, можно указать ParentId для вложенного комментария
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        comment  body  CommentReqCreate  true  "Comment to create"
// @Success      201  {object}  app.Comment  "Created comment"
// @Failure      400  {object}  ErrorResponse  "Invalid input data"
// @Failure      503  {object}  ErrorResponse  "Service unavailable (DB error)"
// @Router       /comments [post]
func (h *CommentHandler) CreateComment(ctx *wbgin.Context) {
	var req CommentReqCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}

	comm, err := h.commentService.CreateComment(req.Text, req.ParentId)

	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, wbgin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, comm)
}

// DeleteComments godoc
// @Summary      Delete Comment
// @Description  Удаляет комментарий и все его дочерние комментарии по ID
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path   string  true  "Comment ID"
// @Success      204  {string}  string  "Comment deleted successfully"
// @Failure      400  {object}  ErrorResponse  "Invalid comment ID"
// @Failure      503  {object}  ErrorResponse  "Service unavailable (DB error)"
// @Router       /comments/{id} [delete]
func (h *CommentHandler) DeleteComments(ctx *wbgin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "id is required"})
		return
	}

	err := h.commentService.DeleteComments(id)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, wbgin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// GetComments godoc
// @Summary      Get Comments
// @Description  Получает комментарии по parentId, поддерживает фильтр search, пагинацию и сортировку
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        parent     query  string  false  "Parent ID (если не указан, можно использовать search)"
// @Param        search     query  string  false  "Текст для поиска комментариев"
// @Param        page       query  int     false  "Номер страницы" default(1)
// @Param        page_size  query  int     false  "Размер страницы" default(10)
// @Param        sort       query  string  false  "Сортировка asc/desc" default(asc)
// @Success      200  {array}   app.CommentNode  "Список комментариев с деревом вложенности"
// @Failure      400  {object}  ErrorResponse    "Invalid parent id"
// @Failure      503  {object}  ErrorResponse    "Service unavailable (DB error)"
// @Router       /comments [get]
func (h *CommentHandler) GetComments(ctx *wbgin.Context) {
	parentId := ctx.Query("parent")
	search := ctx.Query("search")
	page := ctx.Query("page")
	pageSize := ctx.Query("page_size")
	sort := ctx.Query("sort")
	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	var err error
	var nodes []app.CommentNode
	if search == "" {
		nodes, err = h.commentService.GetComments(parentId, sort, pageInt, pageSizeInt)
	} else {
		nodes, err = h.commentService.SearchComments(search, parentId, sort, pageInt, pageSizeInt)
	}
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, wbgin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, nodes)
}

package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	postdomain "backend/internal/domain/post"
	postusecase "backend/internal/usecase/post"

	"github.com/gin-gonic/gin"
)

const (
	messagePostInvalidRequest = "invalid post request"
	messagePostConflict       = "post already exists"
)

// 投稿作成ユースケースの契約。
type CreatePostExecutor interface {
	Execute(ctx context.Context, in *postusecase.CreatePostInput) (*postusecase.CreatePostOutput, error)
}

type PostHandler struct {
	createUsecase CreatePostExecutor
}

// PostHandler を生成する。
func NewPostHandler(usecase CreatePostExecutor) *PostHandler {
	return &PostHandler{createUsecase: usecase}
}

// POST /posts の入力。
type CreatePostRequest struct {
	PostID  string `json:"post_id"`
	Content string `json:"content"`
}

// 作成結果を表す。
type CreatePostResponse struct {
	PostID string `json:"post_id"`
}

/**
 * POST /posts のリクエストを検証し、ユースケースへ委譲して結果を返す。
 */
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req CreatePostRequest
	// JSON パースに失敗したら入力不備
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Message: messagePostInvalidRequest})
		return
	}
	// ID も本文も空は受け付けない
	if strings.TrimSpace(req.PostID) == "" || strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Message: messagePostInvalidRequest})
		return
	}

	out, err := h.createUsecase.Execute(c.Request.Context(), &postusecase.CreatePostInput{
		DarkPostID: req.PostID,
		Content:    req.Content,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreatePostResponse{PostID: out.DarkPostID})
}

/**
 * ユースケースからのエラーを HTTP ステータスとメッセージへ写し替える。
 */
func (h *PostHandler) handleError(c *gin.Context, err error) {
	switch {
	// ユースケースの入力不足
	case errors.Is(err, postusecase.ErrNilInput):
		c.JSON(http.StatusBadRequest, errorResponse{Message: messagePostInvalidRequest})
	// ドメインの空本文エラー
	case errors.Is(err, postdomain.ErrEmptyContent):
		c.JSON(http.StatusBadRequest, errorResponse{Message: messagePostInvalidRequest})
	// 投稿もしくは整形ジョブの重複
	case errors.Is(err, postusecase.ErrPostAlreadyExists),
		errors.Is(err, postusecase.ErrJobAlreadyScheduled):
		c.JSON(http.StatusConflict, errorResponse{Message: messagePostConflict})
	default:
		log.Printf("POST /posts 失敗: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Message: messageInternalError})
	}
}

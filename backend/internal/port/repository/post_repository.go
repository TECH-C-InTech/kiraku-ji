package repository

import (
	"context"
	"errors"

	"backend/internal/domain/post"
)

var (
	ErrPostNotFound      = errors.New("repository: 投稿が見つかりません")
	ErrPostAlreadyExists = errors.New("repository: 投稿がすでに存在します")
)

/**
 * 闇投稿リポジトリの契約
 * Create: 新規保存、重複時は ErrPostAlreadyExists
 * Get: ID 取得、未存在時は ErrPostNotFound
 * ListReady: ready 投稿を最大 limit 件返す
 * Update: 更新、対象欠如時は ErrPostNotFound
 */
type PostRepository interface {
	Create(ctx context.Context, p *post.Post) error
	Get(ctx context.Context, id post.DarkPostID) (*post.Post, error)
	ListReady(ctx context.Context, limit int) ([]*post.Post, error)
	Update(ctx context.Context, p *post.Post) error
}

package repository

import (
	"context"
	"errors"

	"backend/internal/domain/draw"
	"backend/internal/domain/post"
)

var (
	ErrDrawNotFound      = errors.New("repository: おみくじ結果が見つかりません")
	ErrDrawAlreadyExists = errors.New("repository: おみくじ結果がすでに存在します")
)

/**
 * おみくじ結果を扱うリポジトリの契約
 * Create: 新規保存（重複時は ErrDrawAlreadyExists）
 * GetByPostID: 闇投稿 ID から結果を取得（postID が空の場合、未存在時は ErrDrawNotFound）
 * // ListReady: 公開可能（verified）なおみくじ結果一覧を返す。
 */
type DrawRepository interface {
	Create(ctx context.Context, d *draw.Draw) error
	GetByPostID(ctx context.Context, postID post.DarkPostID) (*draw.Draw, error)
	ListReady(ctx context.Context) ([]*draw.Draw, error)
}

package queue

import (
	"context"
	"errors"

	"backend/internal/domain/post"
)

var (
	ErrJobAlreadyScheduled = errors.New("queue: 同一 ID のジョブがすでに存在します")
	ErrQueueClosed         = errors.New("queue: ジョブキューが停止しました")
	ErrContextClosed       = errors.New("queue: コンテキストが終了しました")
)

/**
 * 闇投稿の整形ジョブを溜めたり取り出したりする契約。
 */
type JobQueue interface {
	EnqueueFormat(ctx context.Context, postID post.DarkPostID) error
	DequeueFormat(ctx context.Context) (post.DarkPostID, error)
	Close() error
}

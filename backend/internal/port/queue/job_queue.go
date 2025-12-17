package queue

import (
	"context"
	"errors"

	"backend/internal/domain/post"
)

var (
	ErrJobAlreadyScheduled = errors.New("queue: 同一 ID のジョブがすでに存在します")
)

/**
 * 非同期ジョブキューの契約
 * EnqueueFormat: 闇投稿 ID を LLM に渡す
 */
type JobQueue interface {
	EnqueueFormat(ctx context.Context, postID post.DarkPostID) error
}

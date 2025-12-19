package memory

import (
	"context"
	"fmt"

	"backend/internal/domain/post"
	"backend/internal/port/queue"
)

// チャネルに積み込む簡易ジョブキュー。
type InMemoryJobQueue struct {
	ch chan post.DarkPostID
}

/**
 * 指定バッファでチャネルを用意し、最小 1 件の待ち行列を確保する。
 */
func NewInMemoryJobQueue(buffer int) *InMemoryJobQueue {
	if buffer <= 0 {
		buffer = 1
	}
	return &InMemoryJobQueue{
		ch: make(chan post.DarkPostID, buffer),
	}
}

/**
 * 文脈が閉じられていなければ投稿 ID をチャネルへ積む。
 */
func (q *InMemoryJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", queue.ErrContextClosed, ctx.Err())
	case q.ch <- id:
		return nil
	}
}

/**
 * 文脈が続く限りチャネルから投稿 ID を受け取り、閉鎖済みなら専用エラーを返す。
 */
func (q *InMemoryJobQueue) DequeueFormat(ctx context.Context) (post.DarkPostID, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("%w: %v", queue.ErrContextClosed, ctx.Err())
	case id, ok := <-q.ch:
		if !ok {
			return "", queue.ErrQueueClosed
		}
		return id, nil
	}
}

/**
 * ジョブ供給を止めるためチャネルを閉じる。
 */
func (q *InMemoryJobQueue) Close() error {
	close(q.ch)
	return nil
}

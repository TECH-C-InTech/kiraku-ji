package memory

import (
	"context"

	"backend/internal/domain/post"
	"backend/internal/port/queue"
)

// InMemoryJobQueue はチャネルベースの JobQueue 実装。
type InMemoryJobQueue struct {
	ch chan post.DarkPostID
}

// NewInMemoryJobQueue は指定バッファで JobQueue を生成する。
func NewInMemoryJobQueue(buffer int) *InMemoryJobQueue {
	if buffer <= 0 {
		buffer = 1
	}
	return &InMemoryJobQueue{
		ch: make(chan post.DarkPostID, buffer),
	}
}

func (q *InMemoryJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.ch <- id:
		return nil
	}
}

func (q *InMemoryJobQueue) DequeueFormat(ctx context.Context) (post.DarkPostID, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case id, ok := <-q.ch:
		if !ok {
			return "", queue.ErrQueueClosed
		}
		return id, nil
	}
}

// Close はキューのチャネルを閉じる。
func (q *InMemoryJobQueue) Close() {
	close(q.ch)
}

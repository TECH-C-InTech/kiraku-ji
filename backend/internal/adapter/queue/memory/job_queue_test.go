package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/domain/post"
	"backend/internal/port/queue"
)

func TestInMemoryJobQueue_EnqueueDequeue(t *testing.T) {
	q := NewInMemoryJobQueue(1)
	ctx := context.Background()

	if err := q.EnqueueFormat(ctx, post.DarkPostID("post-1")); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	id, err := q.DequeueFormat(ctx)
	if err != nil {
		t.Fatalf("dequeue failed: %v", err)
	}
	if id != post.DarkPostID("post-1") {
		t.Fatalf("unexpected id: %s", id)
	}

	if err := q.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}
	if _, err := q.DequeueFormat(ctx); !errors.Is(err, queue.ErrQueueClosed) {
		t.Fatalf("expected ErrQueueClosed, got %v", err)
	}
}

func TestInMemoryJobQueue_EnqueueContextCanceled(t *testing.T) {
	q := NewInMemoryJobQueue(1)
	ctx, cancel := context.WithCancel(context.Background())
	if err := q.EnqueueFormat(context.Background(), post.DarkPostID("pre-fill")); err != nil {
		t.Fatalf("failed to pre-fill queue: %v", err)
	}
	cancel()

	err := q.EnqueueFormat(ctx, post.DarkPostID("post-2"))
	if err == nil || !errors.Is(err, queue.ErrContextClosed) {
		t.Fatalf("expected ErrContextClosed, got %v", err)
	}
}

func TestInMemoryJobQueue_DequeueContextCanceled(t *testing.T) {
	q := NewInMemoryJobQueue(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	time.Sleep(time.Nanosecond)
	_, err := q.DequeueFormat(ctx)
	if err == nil || !errors.Is(err, queue.ErrContextClosed) {
		t.Fatalf("expected ErrContextClosed, got %v", err)
	}
}

func TestInMemoryJobQueue_DefaultBuffer(t *testing.T) {
	q := NewInMemoryJobQueue(0)
	if err := q.EnqueueFormat(context.Background(), post.DarkPostID("post-3")); err != nil {
		t.Fatalf("enqueue failed with default buffer: %v", err)
	}
	if _, err := q.DequeueFormat(context.Background()); err != nil {
		t.Fatalf("dequeue failed: %v", err)
	}
}

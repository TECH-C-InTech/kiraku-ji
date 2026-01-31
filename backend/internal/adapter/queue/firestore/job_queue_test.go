package firestore

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"backend/internal/domain/post"
	portqueue "backend/internal/port/queue"
)

func TestFirestoreJobQueue_EnqueueAndDequeue(t *testing.T) {
	queue := newFirestoreJobQueueWithStore(newFakeJobQueueStore())

	ctx := context.Background()
	if err := queue.EnqueueFormat(ctx, post.DarkPostID("post-firestore-1")); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	got, err := queue.DequeueFormat(ctx)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if got != post.DarkPostID("post-firestore-1") {
		t.Fatalf("unexpected id: %s", got)
	}
}

func TestFirestoreJobQueue_DuplicateEnqueueReturnsError(t *testing.T) {
	queue := newFirestoreJobQueueWithStore(newFakeJobQueueStore())

	ctx := context.Background()
	if err := queue.EnqueueFormat(ctx, post.DarkPostID("dup-post")); err != nil {
		t.Fatalf("first enqueue: %v", err)
	}
	if err := queue.EnqueueFormat(ctx, post.DarkPostID("dup-post")); !errors.Is(err, portqueue.ErrJobAlreadyScheduled) {
		t.Fatalf("expected ErrJobAlreadyScheduled, got %v", err)
	}
}

func TestFirestoreJobQueue_DequeueWaitsForNewJob(t *testing.T) {
	queue := newFirestoreJobQueueWithStore(newFakeJobQueueStore())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type result struct {
		id  post.DarkPostID
		err error
	}
	done := make(chan result)
	go func() {
		id, err := queue.DequeueFormat(ctx)
		done <- result{id: id, err: err}
	}()

	time.Sleep(200 * time.Millisecond)
	if err := queue.EnqueueFormat(context.Background(), post.DarkPostID("delayed-post")); err != nil {
		t.Fatalf("enqueue delayed: %v", err)
	}

	select {
	case <-ctx.Done():
		t.Fatalf("context finished before job dequeued: %v", ctx.Err())
	case res := <-done:
		if res.err != nil {
			t.Fatalf("dequeue: %v", res.err)
		}
		if res.id != post.DarkPostID("delayed-post") {
			t.Fatalf("unexpected id: %s", res.id)
		}
	}
}

func TestFirestoreJobQueue_CloseStopsOperations(t *testing.T) {
	queue := newFirestoreJobQueueWithStore(newFakeJobQueueStore())
	if err := queue.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}

	if err := queue.EnqueueFormat(context.Background(), post.DarkPostID("post-after-close")); !errors.Is(err, portqueue.ErrQueueClosed) {
		t.Fatalf("expected ErrQueueClosed on enqueue, got %v", err)
	}

	_, err := queue.DequeueFormat(context.Background())
	if !errors.Is(err, portqueue.ErrQueueClosed) {
		t.Fatalf("expected ErrQueueClosed on dequeue, got %v", err)
	}
}

// Dequeue 中にコンテキストが閉じた場合に適切なエラーへ変換されるかを確認する
func TestFirestoreJobQueue_DequeueContextCanceled(t *testing.T) {
	queue := newFirestoreJobQueueWithStore(newFakeJobQueueStore())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := queue.DequeueFormat(ctx)
	if err == nil || !errors.Is(err, portqueue.ErrContextClosed) {
		t.Fatalf("expected ErrContextClosed, got %v", err)
	}
}

type fakeJobQueueStore struct {
	mu    sync.Mutex
	seq   int64
	items map[post.DarkPostID]int64
}

func newFakeJobQueueStore() *fakeJobQueueStore {
	return &fakeJobQueueStore{
		items: make(map[post.DarkPostID]int64),
	}
}

func (s *fakeJobQueueStore) Create(ctx context.Context, id post.DarkPostID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[id]; exists {
		return portqueue.ErrJobAlreadyScheduled
	}
	s.seq++
	s.items[id] = s.seq
	return nil
}

func (s *fakeJobQueueStore) Dequeue(ctx context.Context) (post.DarkPostID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.items) == 0 {
		return "", errNoJobAvailable
	}
	var (
		selectedID   post.DarkPostID
		selectedSeq int64
		found       bool
	)
	for id, seq := range s.items {
		if !found || seq < selectedSeq {
			selectedID = id
			selectedSeq = seq
			found = true
		}
	}
	delete(s.items, selectedID)
	return selectedID, nil
}

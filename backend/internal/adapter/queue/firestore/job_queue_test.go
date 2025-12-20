package firestore

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"backend/internal/domain/post"
	portqueue "backend/internal/port/queue"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const testProjectID = "firestore-integration-test"

func TestFirestoreJobQueue_EnqueueAndDequeue(t *testing.T) {
	client := newTestFirestoreClient(t)
	truncateCollection(t, client, formatJobsCollection)

	queue, err := NewFirestoreJobQueue(client)
	if err != nil {
		t.Fatalf("NewFirestoreJobQueue: %v", err)
	}

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
	client := newTestFirestoreClient(t)
	truncateCollection(t, client, formatJobsCollection)

	queue, err := NewFirestoreJobQueue(client)
	if err != nil {
		t.Fatalf("NewFirestoreJobQueue: %v", err)
	}

	ctx := context.Background()
	if err := queue.EnqueueFormat(ctx, post.DarkPostID("dup-post")); err != nil {
		t.Fatalf("first enqueue: %v", err)
	}
	if err := queue.EnqueueFormat(ctx, post.DarkPostID("dup-post")); !errors.Is(err, portqueue.ErrJobAlreadyScheduled) {
		t.Fatalf("expected ErrJobAlreadyScheduled, got %v", err)
	}
}

func TestFirestoreJobQueue_DequeueWaitsForNewJob(t *testing.T) {
	client := newTestFirestoreClient(t)
	truncateCollection(t, client, formatJobsCollection)

	queue, err := NewFirestoreJobQueue(client)
	if err != nil {
		t.Fatalf("NewFirestoreJobQueue: %v", err)
	}

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
	client := newTestFirestoreClient(t)
	truncateCollection(t, client, formatJobsCollection)

	queue, err := NewFirestoreJobQueue(client)
	if err != nil {
		t.Fatalf("NewFirestoreJobQueue: %v", err)
	}
	if err := queue.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}

	if err := queue.EnqueueFormat(context.Background(), post.DarkPostID("post-after-close")); !errors.Is(err, portqueue.ErrQueueClosed) {
		t.Fatalf("expected ErrQueueClosed on enqueue, got %v", err)
	}

	_, err = queue.DequeueFormat(context.Background())
	if !errors.Is(err, portqueue.ErrQueueClosed) {
		t.Fatalf("expected ErrQueueClosed on dequeue, got %v", err)
	}
}

// Dequeue 中にコンテキストが閉じた場合に適切なエラーへ変換されるかを確認する
func TestFirestoreJobQueue_DequeueContextCanceled(t *testing.T) {
	client := newTestFirestoreClient(t)
	truncateCollection(t, client, formatJobsCollection)

	queue, err := NewFirestoreJobQueue(client)
	if err != nil {
		t.Fatalf("NewFirestoreJobQueue: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = queue.DequeueFormat(ctx)
	if err == nil || !errors.Is(err, portqueue.ErrContextClosed) {
		t.Fatalf("expected ErrContextClosed, got %v", err)
	}
}

// newTestFirestoreClient は Firestore エミュレータに接続するクライアントを返す。
func newTestFirestoreClient(t *testing.T) *firestore.Client {
	t.Helper()
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST is not set; skipping Firestore queue tests")
	}
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = testProjectID
	}
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("failed to create firestore client: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})
	return client
}

// truncateCollection は指定コレクションを空にする。
func truncateCollection(t *testing.T, client *firestore.Client, collection string) {
	t.Helper()
	ctx := context.Background()
	iter := client.Collection(collection).Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			t.Fatalf("iterate %s: %v", collection, err)
		}
		if _, err := doc.Ref.Delete(ctx); err != nil {
			t.Fatalf("delete doc %s: %v", doc.Ref.ID, err)
		}
	}
}

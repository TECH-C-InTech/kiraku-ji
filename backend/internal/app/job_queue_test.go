package app

import (
	"context"
	"errors"
	"testing"

	queueMemory "backend/internal/adapter/queue/memory"
	"backend/internal/domain/post"
	"backend/internal/port/queue"

	"cloud.google.com/go/firestore"
)

func TestNewJobQueue_DefaultsToMemory(t *testing.T) {
	t.Setenv("JOB_QUEUE_BACKEND", "")
	queue, seed, err := newJobQueue(&Infra{})
	if err != nil {
		t.Fatalf("newJobQueue returned error: %v", err)
	}
	if !seed {
		t.Fatalf("memory backend should enable seeding")
	}
	if _, ok := queue.(*queueMemory.InMemoryJobQueue); !ok {
		t.Fatalf("expected memory job queue, got %T", queue)
	}
}

func TestNewJobQueue_FirestoreRequiresClient(t *testing.T) {
	t.Setenv("JOB_QUEUE_BACKEND", "firestore")
	if _, _, err := newJobQueue(&Infra{}); !errors.Is(err, errFirestoreQueueRequiresClient) {
		t.Fatalf("expected missing client error, got %v", err)
	}
}

func TestNewJobQueue_FirestoreSuccess(t *testing.T) {
	t.Setenv("JOB_QUEUE_BACKEND", "firestore")
	origFactory := firestoreJobQueueFactory
	stub := &fakeJobQueue{}
	firestoreJobQueueFactory = func(client *firestore.Client) (queue.JobQueue, error) {
		return stub, nil
	}
	defer func() { firestoreJobQueueFactory = origFactory }()

	infra := &Infra{firestoreClient: &firestore.Client{}}
	queue, seed, err := newJobQueue(infra)
	if err != nil {
		t.Fatalf("newJobQueue returned error: %v", err)
	}
	if seed {
		t.Fatalf("firestore backend should not request seeding")
	}
	if queue != stub {
		t.Fatalf("expected stub queue instance")
	}
}

func TestNewJobQueue_UnsupportedBackend(t *testing.T) {
	t.Setenv("JOB_QUEUE_BACKEND", "unknown")
	if _, _, err := newJobQueue(&Infra{}); err == nil {
		t.Fatalf("expected error for unsupported backend")
	}
}

type fakeJobQueue struct{}

func (fakeJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	return nil
}

func (fakeJobQueue) DequeueFormat(ctx context.Context) (post.DarkPostID, error) {
	return "", nil
}

func (fakeJobQueue) Close() error {
	return nil
}

var _ queue.JobQueue = (*fakeJobQueue)(nil)

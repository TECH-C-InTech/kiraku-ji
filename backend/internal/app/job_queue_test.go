package app

import (
	"context"
	"errors"
	"testing"

	"backend/internal/domain/post"
	"backend/internal/port/queue"

	"cloud.google.com/go/firestore"
)

func TestNewJobQueue_RequiresFirestoreClient(t *testing.T) {
	if _, err := newJobQueue(&Infra{}); !errors.Is(err, errFirestoreQueueRequiresClient) {
		t.Fatalf("expected missing client error, got %v", err)
	}
}

func TestNewJobQueue_FirestoreSuccess(t *testing.T) {
	origFactory := firestoreJobQueueFactory
	stub := &fakeJobQueue{}
	firestoreJobQueueFactory = func(client *firestore.Client) (queue.JobQueue, error) {
		return stub, nil
	}
	defer func() { firestoreJobQueueFactory = origFactory }()

	infra := &Infra{firestoreClient: &firestore.Client{}}
	queue, err := newJobQueue(infra)
	if err != nil {
		t.Fatalf("newJobQueue returned error: %v", err)
	}
	if queue != stub {
		t.Fatalf("expected stub queue instance")
	}
}

func TestNewJobQueue_FactoryError(t *testing.T) {
	origFactory := firestoreJobQueueFactory
	defer func() { firestoreJobQueueFactory = origFactory }()

	firestoreJobQueueFactory = func(client *firestore.Client) (queue.JobQueue, error) {
		return nil, errors.New("factory error")
	}

	infra := &Infra{firestoreClient: &firestore.Client{}}
	if _, err := newJobQueue(infra); err == nil {
		t.Fatalf("expected error when queue factory fails")
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

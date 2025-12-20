package app

import (
	"context"
	"testing"

	"backend/internal/domain/post"
	"backend/internal/port/repository"

	"cloud.google.com/go/firestore"
)

func TestNewAPIPostRepository_FailsWithoutFirestoreClient(t *testing.T) {
	t.Parallel()
	repo, err := newAPIPostRepository(&Infra{})
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if repo != nil {
		t.Fatalf("expected repo to be nil when Firestore client is missing")
	}
	if err != errFirestoreClientUnavailable {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewAPIPostRepository_Success(t *testing.T) {
	t.Parallel()
	originalFactory := apiPostRepositoryFactory
	t.Cleanup(func() {
		apiPostRepositoryFactory = originalFactory
	})

	mockClient := &firestore.Client{}
	expectedRepo := &stubPostRepository{}
	apiPostRepositoryFactory = func(client *firestore.Client) (repository.PostRepository, error) {
		if client != mockClient {
			t.Fatalf("unexpected client")
		}
		return expectedRepo, nil
	}

	infra := &Infra{firestoreClient: mockClient}
	repo, err := newAPIPostRepository(infra)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo != expectedRepo {
		t.Fatalf("unexpected repository instance")
	}
}

type stubPostRepository struct{}

func (stubPostRepository) Create(context.Context, *post.Post) error {
	return nil
}

func (stubPostRepository) Get(context.Context, post.DarkPostID) (*post.Post, error) {
	return nil, nil
}

func (stubPostRepository) ListReady(context.Context, int) ([]*post.Post, error) {
	return nil, nil
}

func (stubPostRepository) Update(context.Context, *post.Post) error {
	return nil
}

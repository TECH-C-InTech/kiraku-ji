package memory

import (
	"context"
	"errors"
	"testing"

	"backend/internal/domain/post"
	"backend/internal/port/repository"
)

func TestInMemoryPostRepository_CreateAndGet(t *testing.T) {
	repo := NewInMemoryPostRepository()
	p, err := post.New("post-1", "content")
	if err != nil {
		t.Fatalf("failed to create domain post: %v", err)
	}

	if err := repo.Create(context.Background(), p); err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	got, err := repo.Get(context.Background(), p.ID())
	if err != nil {
		t.Fatalf("get returned error: %v", err)
	}
	if got.ID() != p.ID() {
		t.Fatalf("unexpected post id: %s", got.ID())
	}
}

func TestInMemoryPostRepository_CreateDuplicate(t *testing.T) {
	repo := NewInMemoryPostRepository()
	p, _ := post.New("post-1", "content")
	_ = repo.Create(context.Background(), p)

	err := repo.Create(context.Background(), p)
	if !errors.Is(err, repository.ErrPostAlreadyExists) {
		t.Fatalf("expected ErrPostAlreadyExists, got %v", err)
	}
}

func TestInMemoryPostRepository_GetNotFound(t *testing.T) {
	repo := NewInMemoryPostRepository()
	if _, err := repo.Get(context.Background(), post.DarkPostID("missing")); !errors.Is(err, repository.ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestInMemoryPostRepository_ListReadyAndLimit(t *testing.T) {
	repo := NewInMemoryPostRepository()
	ready, _ := post.New("post-ready", "ready content")
	if err := ready.MarkReady(); err != nil {
		t.Fatalf("mark ready failed: %v", err)
	}
	pending, _ := post.New("post-pending", "pending content")

	_ = repo.Create(context.Background(), ready)
	_ = repo.Create(context.Background(), pending)

	list, err := repo.ListReady(context.Background(), 1)
	if err != nil {
		t.Fatalf("list ready returned error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 ready post, got %d", len(list))
	}
	if list[0].ID() != ready.ID() {
		t.Fatalf("unexpected ready post: %s", list[0].ID())
	}
}

func TestInMemoryPostRepository_Update(t *testing.T) {
	repo := NewInMemoryPostRepository()
	p, _ := post.New("post-1", "content")
	if err := repo.Create(context.Background(), p); err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	if err := repo.Update(context.Background(), p); err != nil {
		t.Fatalf("update returned error: %v", err)
	}

	newPost, _ := post.New("missing", "content")
	err := repo.Update(context.Background(), newPost)
	if !errors.Is(err, repository.ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

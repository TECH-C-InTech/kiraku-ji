package memory

import (
	"context"
	"errors"
	"testing"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/repository"
)

func TestInMemoryDrawRepository_CreateAndGet(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryDrawRepository()
	ctx := context.Background()

	draw := newVerifiedDraw(t, "post-1", "fortune-1")
	if err := repo.Create(ctx, draw); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 同じ ID の Create は ErrDrawAlreadyExists
	if err := repo.Create(ctx, draw); !errors.Is(err, repository.ErrDrawAlreadyExists) {
		t.Fatalf("expected ErrDrawAlreadyExists, got %v", err)
	}

	got, err := repo.GetByPostID(ctx, draw.PostID())
	if err != nil {
		t.Fatalf("GetByPostID() error = %v", err)
	}
	if got.PostID() != draw.PostID() {
		t.Fatalf("unexpected post id: want %s, got %s", draw.PostID(), got.PostID())
	}

	// 存在しない ID は ErrDrawNotFound
	if _, err := repo.GetByPostID(ctx, post.DarkPostID("post-x")); !errors.Is(err, repository.ErrDrawNotFound) {
		t.Fatalf("expected ErrDrawNotFound, got %v", err)
	}
}

func TestInMemoryDrawRepository_ListReady(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryDrawRepository()
	ctx := context.Background()

	verified := newVerifiedDraw(t, "post-ready", "ready")
	if err := repo.Create(ctx, verified); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// pending な draw
	pending, err := drawdomain.New(post.DarkPostID("post-pending"), drawdomain.FormattedContent("pending"))
	if err != nil {
		t.Fatalf("drawdomain.New() error = %v", err)
	}
	if err := repo.Create(ctx, pending); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	results, err := repo.ListReady(ctx)
	if err != nil {
		t.Fatalf("ListReady() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 verified draw, got %d", len(results))
	}
	if results[0].PostID() != verified.PostID() {
		t.Fatalf("unexpected draw returned: want %s, got %s", verified.PostID(), results[0].PostID())
	}
}

func newVerifiedDraw(t *testing.T, postID, result string) *drawdomain.Draw {
	t.Helper()

	d, err := drawdomain.New(post.DarkPostID(postID), drawdomain.FormattedContent(result))
	if err != nil {
		t.Fatalf("drawdomain.New() error = %v", err)
	}
	d.MarkVerified()
	return d
}

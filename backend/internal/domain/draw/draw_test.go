package draw

import (
	"testing"

	"backend/internal/domain/post"
)

func TestNew(t *testing.T) {
	t.Parallel()

	draw, err := New(post.DarkPostID("post-id"), FormattedContent("やさしい言葉"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if draw.PostID() != post.DarkPostID("post-id") {
		t.Fatalf("unexpected post id: %s", string(draw.PostID()))
	}
	if draw.Result() != FormattedContent("やさしい言葉") {
		t.Fatalf("unexpected result: %s", string(draw.Result()))
	}
	if draw.Status() != StatusPending {
		t.Fatalf("expected status pending but got %s", draw.Status())
	}
}

func TestNew_EmptyResult(t *testing.T) {
	t.Parallel()

	if _, err := New(post.DarkPostID("post-id"), FormattedContent("")); err != ErrEmptyResult {
		t.Fatalf("expected ErrEmptyResult but got %v", err)
	}
}

func TestFromPost(t *testing.T) {
	t.Parallel()

	p, err := post.New(post.DarkPostID("post-id"), post.DarkContent("闇が深い"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := p.MarkReady(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	draw, err := FromPost(p, FormattedContent("やさしい言葉"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if draw.PostID() != p.ID() {
		t.Fatalf("expected post id %s but got %s", p.ID(), draw.PostID())
	}
}

func TestFromPost_NotReady(t *testing.T) {
	t.Parallel()

	p, err := post.New(post.DarkPostID("post-id"), post.DarkContent("闇"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := FromPost(p, FormattedContent("まだ早い")); err != ErrPostNotReady {
		t.Fatalf("expected ErrPostNotReady but got %v", err)
	}
}

func TestMarkVerified(t *testing.T) {
	t.Parallel()

	draw, err := New(post.DarkPostID("post-id"), FormattedContent("result"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	draw.MarkVerified()
	if draw.Status() != StatusVerified {
		t.Fatalf("expected status verified but got %s", draw.Status())
	}
}

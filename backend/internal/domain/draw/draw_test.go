package draw

import (
	"testing"

	"backend/internal/domain/post"
)

func TestNew(t *testing.T) {
	t.Parallel()

	draw, err := New(post.DarkPostID("post-id"), "やさしい言葉")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if draw.PostID() != post.DarkPostID("post-id") {
		t.Fatalf("unexpected post id: %s", string(draw.PostID()))
	}
	if draw.Result() != "やさしい言葉" {
		t.Fatalf("unexpected result: %s", draw.Result())
	}
}

func TestNew_EmptyResult(t *testing.T) {
	t.Parallel()

	if _, err := New(post.DarkPostID("post-id"), ""); err != ErrEmptyResult {
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

	draw, err := FromPost(p, "やさしい言葉")
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

	if _, err := FromPost(p, "まだ早い"); err != ErrPostNotReady {
		t.Fatalf("expected ErrPostNotReady but got %v", err)
	}
}

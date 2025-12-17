package post

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	post, err := New(DarkPostID("post-id"), DarkContent("闇がおおい"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.ID() != DarkPostID("post-id") {
		t.Fatalf("unexpected id: %s", string(post.ID()))
	}
	if post.Content() != DarkContent("闇がおおい") {
		t.Fatalf("unexpected content: %s", string(post.Content()))
	}
	if post.Status() != StatusPending {
		t.Fatalf("expected pending but got %s", post.Status())
	}
}

func TestNew_EmptyContent(t *testing.T) {
	t.Parallel()

	if _, err := New(DarkPostID("post-id"), DarkContent("")); err != ErrEmptyContent {
		t.Fatalf("expected ErrEmptyContent but got %v", err)
	}
}

func TestMarkReady(t *testing.T) {
	t.Parallel()

	post, err := New(DarkPostID("id"), DarkContent("闇"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := post.MarkReady(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.Status() != StatusReady {
		t.Fatalf("expected ready but got %s", post.Status())
	}
}

func TestMarkReady_InvalidTransition(t *testing.T) {
	t.Parallel()

	post, err := Restore(DarkPostID("id"), DarkContent("闇"), StatusReady)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := post.MarkReady(); err != ErrInvalidStatusTransition {
		t.Fatalf("expected ErrInvalidStatusTransition but got %v", err)
	}
}

func TestRestore_InvalidStatus(t *testing.T) {
	t.Parallel()

	if _, err := Restore(DarkPostID("id"), DarkContent("闇"), Status("unknown")); err != ErrInvalidStatus {
		t.Fatalf("expected ErrInvalidStatus but got %v", err)
	}
}

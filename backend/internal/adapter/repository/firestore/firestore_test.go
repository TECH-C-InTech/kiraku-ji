package firestore

import (
	"context"
	"testing"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/repository"
)

func TestDrawRepository_CreateGetListReady(t *testing.T) {
	repo := &DrawRepository{store: newFakeDrawStore()}

	ctx := context.Background()
	draw, err := drawdomain.New(post.DarkPostID("post-1"), drawdomain.FormattedContent("fortune smiles"))
	if err != nil {
		t.Fatalf("new draw: %v", err)
	}
	draw.MarkVerified()

	if err := repo.Create(ctx, draw); err != nil {
		t.Fatalf("create draw: %v", err)
	}

	fetched, err := repo.GetByPostID(ctx, "post-1")
	if err != nil {
		t.Fatalf("get draw: %v", err)
	}
	if fetched.Result() != draw.Result() || fetched.Status() != draw.Status() {
		t.Fatalf("fetched draw mismatch")
	}

	list, err := repo.ListReady(ctx)
	if err != nil {
		t.Fatalf("list ready: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 draw got %d", len(list))
	}
}

type fakeDrawStore struct {
	items map[post.DarkPostID]drawDocument
}

func newFakeDrawStore() *fakeDrawStore {
	return &fakeDrawStore{
		items: make(map[post.DarkPostID]drawDocument),
	}
}

func (s *fakeDrawStore) Create(ctx context.Context, doc drawDocument) error {
	id := post.DarkPostID(doc.PostID)
	if _, exists := s.items[id]; exists {
		return repository.ErrDrawAlreadyExists
	}
	s.items[id] = doc
	return nil
}

func (s *fakeDrawStore) Get(ctx context.Context, postID post.DarkPostID) (drawDocument, error) {
	doc, ok := s.items[postID]
	if !ok {
		return drawDocument{}, repository.ErrDrawNotFound
	}
	return doc, nil
}

func (s *fakeDrawStore) ListReady(ctx context.Context) ([]drawDocument, error) {
	var documents []drawDocument
	for _, doc := range s.items {
		if doc.Status == string(drawdomain.StatusVerified) {
			documents = append(documents, doc)
		}
	}
	return documents, nil
}

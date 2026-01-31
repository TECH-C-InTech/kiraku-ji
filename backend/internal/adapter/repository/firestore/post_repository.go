package firestore

import (
	"context"
	"errors"
	"fmt"

	postdomain "backend/internal/domain/post"
	"backend/internal/port/repository"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// postsCollection は Firestore 上の posts コレクション名。
const postsCollection = "posts"

var (
	// errNilPost は nil を保存しようとした際のバリデーションエラー。
	errNilPost = errors.New("firestorerepository: post is nil")
	// errMissingClient は Firestore クライアント未設定時の初期化エラー。
	errMissingClient = errors.New("firestorerepository: firestore client is missing")
)

// postDocument は Firestore の posts ドキュメント構造を表す。
type postDocument struct {
	PostID  string `firestore:"post_id"`
	Content string `firestore:"content"`
	Status  string `firestore:"status"`
}

type postStore interface {
	Create(ctx context.Context, doc postDocument) error
	Get(ctx context.Context, id postdomain.DarkPostID) (postDocument, error)
	ListReady(ctx context.Context, limit int) ([]postDocument, error)
	Update(ctx context.Context, doc postDocument) error
}

type firestorePostStore struct {
	client *firestore.Client
}

// PostRepository は Firestore を利用した Post リポジトリ実装。
type PostRepository struct {
	store postStore
}

// NewPostRepository は Firestore クライアントを受け取って PostRepository を作成する。
func NewPostRepository(client *firestore.Client) (*PostRepository, error) {
	store, err := newFirestorePostStore(client)
	if err != nil {
		return nil, err
	}
	return &PostRepository{store: store}, nil
}

// Create は新しい Post を Firestore に保存する。
func (r *PostRepository) Create(ctx context.Context, p *postdomain.Post) error {
	if p == nil {
		return errNilPost
	}

	doc := postDocument{
		PostID:  string(p.ID()),
		Content: string(p.Content()),
		Status:  string(p.Status()),
	}
	return r.store.Create(ctx, doc)
}

// Get は指定 ID の Post を Firestore から取得する。
func (r *PostRepository) Get(ctx context.Context, id postdomain.DarkPostID) (*postdomain.Post, error) {
	if id == "" {
		return nil, repository.ErrPostNotFound
	}

	doc, err := r.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return restorePostFromDocument(doc)
}

// ListReady は ready 状態の Post を最大 limit 件取得する。
func (r *PostRepository) ListReady(ctx context.Context, limit int) ([]*postdomain.Post, error) {
	documents, err := r.store.ListReady(ctx, limit)
	if err != nil {
		return nil, err
	}
	posts := make([]*postdomain.Post, 0, len(documents))
	for _, doc := range documents {
		p, err := restorePostFromDocument(doc)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

// Update は既存の Post を Firestore 上で更新する。
func (r *PostRepository) Update(ctx context.Context, p *postdomain.Post) error {
	if p == nil {
		return errNilPost
	}
	if p.ID() == "" {
		return repository.ErrPostNotFound
	}

	doc := postDocument{
		PostID:  string(p.ID()),
		Content: string(p.Content()),
		Status:  string(p.Status()),
	}
	return r.store.Update(ctx, doc)
}

// restorePostFromDocument は Post の保存データからドメインを復元する。
func restorePostFromDocument(payload postDocument) (*postdomain.Post, error) {
	post, err := postdomain.Restore(
		postdomain.DarkPostID(payload.PostID),
		postdomain.DarkContent(payload.Content),
		postdomain.Status(payload.Status),
	)
	if err != nil {
		return nil, fmt.Errorf("restore post: %w", err)
	}
	return post, nil
}

func newFirestorePostStore(client *firestore.Client) (*firestorePostStore, error) {
	if client == nil {
		return nil, errMissingClient
	}
	return &firestorePostStore{client: client}, nil
}

func (s *firestorePostStore) Create(ctx context.Context, doc postDocument) error {
	ref := s.client.Collection(postsCollection).Doc(doc.PostID)
	data := map[string]any{
		"post_id":    doc.PostID,
		"content":    doc.Content,
		"status":     doc.Status,
		"created_at": firestore.ServerTimestamp,
	}
	_, err := ref.Create(ctx, data)
	if status.Code(err) == codes.AlreadyExists {
		return repository.ErrPostAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create post document: %w", err)
	}
	return nil
}

func (s *firestorePostStore) Get(ctx context.Context, id postdomain.DarkPostID) (postDocument, error) {
	doc, err := s.client.Collection(postsCollection).Doc(string(id)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return postDocument{}, repository.ErrPostNotFound
		}
		return postDocument{}, fmt.Errorf("get post document: %w", err)
	}
	var payload postDocument
	if err := doc.DataTo(&payload); err != nil {
		return postDocument{}, fmt.Errorf("decode post document: %w", err)
	}
	return payload, nil
}

func (s *firestorePostStore) ListReady(ctx context.Context, limit int) ([]postDocument, error) {
	query := s.client.Collection(postsCollection).
		Where("status", "==", string(postdomain.StatusReady))
	if limit > 0 {
		query = query.Limit(limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var documents []postDocument
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterate ready posts: %w", err)
		}
		var payload postDocument
		if err := doc.DataTo(&payload); err != nil {
			return nil, fmt.Errorf("decode post document: %w", err)
		}
		documents = append(documents, payload)
	}
	return documents, nil
}

func (s *firestorePostStore) Update(ctx context.Context, doc postDocument) error {
	ref := s.client.Collection(postsCollection).Doc(doc.PostID)
	updates := []firestore.Update{
		{Path: "content", Value: doc.Content},
		{Path: "status", Value: doc.Status},
		{Path: "updated_at", Value: firestore.ServerTimestamp},
	}

	_, err := ref.Update(ctx, updates)
	if status.Code(err) == codes.NotFound {
		return repository.ErrPostNotFound
	}
	if err != nil {
		return fmt.Errorf("update post document: %w", err)
	}
	return nil
}

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

// PostRepository は Firestore を利用した Post リポジトリ実装。
type PostRepository struct {
	client *firestore.Client
}

// NewPostRepository は Firestore クライアントを受け取って PostRepository を作成する。
func NewPostRepository(client *firestore.Client) (*PostRepository, error) {
	if client == nil {
		return nil, errMissingClient
	}
	return &PostRepository{client: client}, nil
}

// Create は新しい Post を Firestore に保存する。
func (r *PostRepository) Create(ctx context.Context, p *postdomain.Post) error {
	if p == nil {
		return errNilPost
	}

	doc := r.client.Collection(postsCollection).Doc(string(p.ID()))
	data := map[string]any{
		"post_id":    string(p.ID()),
		"content":    string(p.Content()),
		"status":     string(p.Status()),
		"created_at": firestore.ServerTimestamp,
	}

	_, err := doc.Create(ctx, data)
	if status.Code(err) == codes.AlreadyExists {
		return repository.ErrPostAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create post document: %w", err)
	}
	return nil
}

// Get は指定 ID の Post を Firestore から取得する。
func (r *PostRepository) Get(ctx context.Context, id postdomain.DarkPostID) (*postdomain.Post, error) {
	if id == "" {
		return nil, repository.ErrPostNotFound
	}

	doc, err := r.client.Collection(postsCollection).Doc(string(id)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, repository.ErrPostNotFound
		}
		return nil, fmt.Errorf("get post document: %w", err)
	}

	return restorePostFromDoc(doc)
}

// ListReady は ready 状態の Post を最大 limit 件取得する。
func (r *PostRepository) ListReady(ctx context.Context, limit int) ([]*postdomain.Post, error) {
	query := r.client.Collection(postsCollection).
		Where("status", "==", string(postdomain.StatusReady))
	if limit > 0 {
		query = query.Limit(limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var posts []*postdomain.Post
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterate ready posts: %w", err)
		}

		p, err := restorePostFromDoc(doc)
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

	doc := r.client.Collection(postsCollection).Doc(string(p.ID()))
	updates := []firestore.Update{
		{Path: "content", Value: string(p.Content())},
		{Path: "status", Value: string(p.Status())},
		{Path: "updated_at", Value: firestore.ServerTimestamp},
	}

	_, err := doc.Update(ctx, updates)
	if status.Code(err) == codes.NotFound {
		return repository.ErrPostNotFound
	}
	if err != nil {
		return fmt.Errorf("update post document: %w", err)
	}
	return nil
}

// restorePostFromDoc は Firestore ドキュメントから Post ドメインを復元する。
func restorePostFromDoc(doc *firestore.DocumentSnapshot) (*postdomain.Post, error) {
	var payload postDocument
	if err := doc.DataTo(&payload); err != nil {
		return nil, fmt.Errorf("decode post document: %w", err)
	}

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

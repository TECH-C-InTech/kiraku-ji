package firestore

import (
	"context"
	"errors"
	"fmt"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/repository"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// drawsCollection は Firestore 上のコレクション名。
const drawsCollection = "draws"

var (
	// errNilDraw は nil を保存しようとした際のバリデーションエラー。
	errNilDraw = errors.New("firestorerepository: draw is nil")
	// errEmptyPostID は ID が空のまま操作した際のエラー。
	errEmptyPostID = errors.New("firestorerepository: post id is empty")
	// errMissingRepository は Firestore クライアント欠如時の初期化エラー。
	errMissingRepository = errors.New("firestorerepository: firestore client is missing")
)

// ポインタですよということでfirestoreの接続などが実行できる
type DrawRepository struct {
	store drawStore
}

type drawDocument struct {
	PostID string `firestore:"post_id"`
	Result string `firestore:"result"`
	Status string `firestore:"status"`
}

type drawStore interface {
	Create(ctx context.Context, doc drawDocument) error
	Get(ctx context.Context, postID post.DarkPostID) (drawDocument, error)
	ListReady(ctx context.Context) ([]drawDocument, error)
}

type firestoreDrawStore struct {
	client *firestore.Client
}

// NewDrawRepository は Firestore を利用するリポジトリを生成する。
func NewDrawRepository(client *firestore.Client) (*DrawRepository, error) {
	store, err := newFirestoreDrawStore(client)
	if err != nil {
		return nil, err
	}
	return &DrawRepository{store: store}, nil
}

// Create は Draw を Firestore に保存する。
func (r *DrawRepository) Create(ctx context.Context, d *drawdomain.Draw) error {
	if d == nil {
		return errNilDraw
	}
	postID := d.PostID()
	if postID == "" {
		return errEmptyPostID
	}

	doc := drawDocument{
		PostID: string(d.PostID()),
		Result: string(d.Result()),
		Status: string(d.Status()),
	}
	return r.store.Create(ctx, doc)
}

// GetByPostID は Firestore から Draw を取得する。
func (r *DrawRepository) GetByPostID(ctx context.Context, postID post.DarkPostID) (*drawdomain.Draw, error) {
	if postID == "" {
		return nil, repository.ErrDrawNotFound
	}

	doc, err := r.store.Get(ctx, postID)
	if err != nil {
		return nil, err
	}

	return restoreDrawFromDocument(doc)
}

// ListReady は Verified な Draw を Firestore から列挙する。
func (r *DrawRepository) ListReady(ctx context.Context) ([]*drawdomain.Draw, error) {
	documents, err := r.store.ListReady(ctx)
	if err != nil {
		return nil, err
	}
	draws := make([]*drawdomain.Draw, 0, len(documents))
	for _, doc := range documents {
		d, err := restoreDrawFromDocument(doc)
		if err != nil {
			return nil, err
		}
		draws = append(draws, d)
	}
	return draws, nil
}

// restoreDrawFromDocument は Draw の保存データからドメインを復元する。
func restoreDrawFromDocument(payload drawDocument) (*drawdomain.Draw, error) {
	restored, err := drawdomain.Restore(
		post.DarkPostID(payload.PostID),
		drawdomain.FormattedContent(payload.Result),
		drawdomain.Status(payload.Status),
	)
	if err != nil {
		return nil, fmt.Errorf("restore draw: %w", err)
	}
	return restored, nil
}

func newFirestoreDrawStore(client *firestore.Client) (*firestoreDrawStore, error) {
	if client == nil {
		return nil, errMissingRepository
	}
	return &firestoreDrawStore{client: client}, nil
}

func (s *firestoreDrawStore) Create(ctx context.Context, doc drawDocument) error {
	ref := s.client.Collection(drawsCollection).Doc(doc.PostID)
	data := map[string]interface{}{
		"post_id":    doc.PostID,
		"result":     doc.Result,
		"status":     doc.Status,
		"created_at": firestore.ServerTimestamp,
	}
	_, err := ref.Create(ctx, data)
	if status.Code(err) == codes.AlreadyExists {
		return repository.ErrDrawAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create draw document: %w", err)
	}
	return nil
}

func (s *firestoreDrawStore) Get(ctx context.Context, postID post.DarkPostID) (drawDocument, error) {
	doc, err := s.client.Collection(drawsCollection).Doc(string(postID)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return drawDocument{}, repository.ErrDrawNotFound
		}
		return drawDocument{}, fmt.Errorf("get draw document: %w", err)
	}
	var payload drawDocument
	if err := doc.DataTo(&payload); err != nil {
		return drawDocument{}, fmt.Errorf("decode draw document: %w", err)
	}
	return payload, nil
}

func (s *firestoreDrawStore) ListReady(ctx context.Context) ([]drawDocument, error) {
	iter := s.client.Collection(drawsCollection).
		Where("status", "==", string(drawdomain.StatusVerified)).
		Documents(ctx)
	defer iter.Stop()

	var documents []drawDocument
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterate verified draws: %w", err)
		}
		var payload drawDocument
		if err := doc.DataTo(&payload); err != nil {
			return nil, fmt.Errorf("decode draw document: %w", err)
		}
		documents = append(documents, payload)
	}
	return documents, nil
}

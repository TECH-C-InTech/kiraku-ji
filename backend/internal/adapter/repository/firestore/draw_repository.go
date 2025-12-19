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
	client *firestore.Client
}

// NewDrawRepository は Firestore を利用するリポジトリを生成する。
func NewDrawRepository(client *firestore.Client) (*DrawRepository, error) {
	if client == nil {
		return nil, errMissingRepository
	}
	return &DrawRepository{client: client}, nil
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

	doc := r.client.Collection(drawsCollection).Doc(string(postID))
	// Firestore に保存するフィールド群。
	data := map[string]interface{}{
		"post_id":    string(d.PostID()),
		"result":     string(d.Result()),
		"status":     string(d.Status()),
		"created_at": firestore.ServerTimestamp,
	}

	//保存するときのエラーチェック
	_, err := doc.Create(ctx, data)
	if status.Code(err) == codes.AlreadyExists {
		return repository.ErrDrawAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create draw document: %w", err)
	}
	return nil
}

// GetByPostID は Firestore から Draw を取得する。
func (r *DrawRepository) GetByPostID(ctx context.Context, postID post.DarkPostID) (*drawdomain.Draw, error) {
	if postID == "" {
		return nil, repository.ErrDrawNotFound
	}

	doc, err := r.client.Collection(drawsCollection).Doc(string(postID)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, repository.ErrDrawNotFound
		}
		return nil, fmt.Errorf("get draw document: %w", err)
	}

	return restoreDrawFromDoc(doc)
}

// ListReady は Verified な Draw を Firestore から列挙する。
func (r *DrawRepository) ListReady(ctx context.Context) ([]*drawdomain.Draw, error) {
	// status が verified の個体のみ抽出するクエリ。
	iter := r.client.Collection(drawsCollection).
		Where("status", "==", string(drawdomain.StatusVerified)).
		Documents(ctx)
	defer iter.Stop()

	var draws []*drawdomain.Draw
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterate verified draws: %w", err)
		}

		d, err := restoreDrawFromDoc(doc)
		if err != nil {
			return nil, err
		}
		draws = append(draws, d)
	}

	return draws, nil
}

// restoreDrawFromDoc は Firestore ドキュメントをドメインオブジェクトに変換する。
func restoreDrawFromDoc(doc *firestore.DocumentSnapshot) (*drawdomain.Draw, error) {
	var payload struct {
		PostID string `firestore:"post_id"`
		Result string `firestore:"result"`
		Status string `firestore:"status"`
	}
	if err := doc.DataTo(&payload); err != nil {
		return nil, fmt.Errorf("decode draw document: %w", err)
	}

	restored, err := drawdomain.Restore(post.DarkPostID(payload.PostID), drawdomain.FormattedContent(payload.Result), drawdomain.Status(payload.Status))
	if err != nil {
		return nil, fmt.Errorf("restore draw: %w", err)
	}
	return restored, nil
}

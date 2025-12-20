package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"backend/internal/adapter/http/handler"
	firestoreadapter "backend/internal/adapter/repository/firestore"
	memoryrepo "backend/internal/adapter/repository/memory"
	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/repository"
	drawusecase "backend/internal/usecase/draw"
	postusecase "backend/internal/usecase/post"
	"cloud.google.com/go/firestore"
)

// Container は API で使用する依存を保持する。
type Container struct {
	Infra              *Infra
	DrawFortuneUsecase *drawusecase.FortuneUsecase
	DrawHandler        *handler.DrawHandler
	CreatePostUsecase  *postusecase.CreatePostUsecase
	PostHandler        *handler.PostHandler
}

// NewContainer は依存を初期化して返す。
func NewContainer(ctx context.Context) (*Container, error) {
	infra, err := NewInfra(ctx)
	if err != nil {
		return nil, fmt.Errorf("init infra: %w", err)
	}

	repo, err := provideDrawRepository(ctx, infra)
	if err != nil {
		return nil, fmt.Errorf("provide draw repository: %w", err)
	}

	usecase := drawusecase.NewFortuneUsecase(repo)
	drawHandler := handler.NewDrawHandler(usecase)

	// API では Firestore へ統一するため、メモリ実装へは切り替えない
	postRepo, err := newAPIPostRepository(infra)
	if err != nil {
		return nil, fmt.Errorf("init post repository: %w", err)
	}
	// 投稿整形キューは Firestore 固定の format_jobs を利用する
	jobQueue, err := jobQueueFactory(infra)
	if err != nil {
		return nil, fmt.Errorf("init job queue: %w", err)
	}
	createPostUsecase := postusecase.NewCreatePostUsecase(postRepo, jobQueue)
	postHandler := handler.NewPostHandler(createPostUsecase)

	return &Container{
		Infra:              infra,
		DrawFortuneUsecase: usecase,
		DrawHandler:        drawHandler,
		CreatePostUsecase:  createPostUsecase,
		PostHandler:        postHandler,
	}, nil
}

// Close は保持している外部リソースをクローズする。
func (c *Container) Close() error {
	if c == nil || c.Infra == nil {
		return nil
	}
	return c.Infra.Close()
}

var (
	errFirestoreClientUnavailable = errors.New("post repository: Firestore クライアントが初期化されていません")
	apiPostRepositoryFactory      = func(client *firestore.Client) (repository.PostRepository, error) {
		return firestoreadapter.NewPostRepository(client)
	}
)

/**
 * API 用に Firestore 固定の投稿リポジトリを構築する。
 */
func newAPIPostRepository(infra *Infra) (repository.PostRepository, error) {
	client := infra.Firestore()
	if client == nil {
		return nil, errFirestoreClientUnavailable
	}
	// PostRepository も Firestore で統一し、クライアント生成失敗を上位へ伝播させる
	repo, err := apiPostRepositoryFactory(client)
	if err != nil {
		return nil, fmt.Errorf("new firestore post repository: %w", err)
	}
	return repo, nil
}

func provideDrawRepository(ctx context.Context, infra *Infra) (repository.DrawRepository, error) {
	mode := os.Getenv("DRAW_REPOSITORY_MODE")
	if mode == "error" {
		return newFailingDrawRepository(), nil
	}

	if mode == "firestore" {
		return newFirestoreDrawRepository(infra)
	}

	repo := memoryrepo.NewInMemoryDrawRepository()
	if err := seedDraws(ctx, repo, mode); err != nil {
		return nil, err
	}
	return repo, nil
}

func newFirestoreDrawRepository(infra *Infra) (repository.DrawRepository, error) {
	client := infra.Firestore()
	if client == nil {
		return nil, errors.New("firestore draw repository requested but firestore client is not initialized")
	}
	// Firestore 実装を初期化し、アダプター層経由で返す。
	return firestoreadapter.NewDrawRepository(client)
}

func seedDraws(ctx context.Context, repo repository.DrawRepository, mode string) error {
	samples := []struct {
		id      string
		content string
		ready   bool
	}{
		{id: "post-verified", content: "すべてはうまくいくでしょう", ready: true},
		{id: "post-pending", content: "しばらく待つと吉", ready: false},
	}

	if mode == "empty" {
		for i := range samples {
			samples[i].ready = false
		}
	}

	for _, sample := range samples {
		draw, err := drawdomain.New(post.DarkPostID(sample.id), drawdomain.FormattedContent(sample.content))
		if err != nil {
			return err
		}
		if sample.ready {
			draw.MarkVerified()
		}
		if err := repo.Create(ctx, draw); err != nil && err != repository.ErrDrawAlreadyExists {
			return err
		}
	}
	return nil
}

type failingDrawRepository struct {
	err error
}

func newFailingDrawRepository() repository.DrawRepository {
	return &failingDrawRepository{
		err: errors.New("forced repository error (DRAW_REPOSITORY_MODE=error)"),
	}
}

func (f *failingDrawRepository) Create(ctx context.Context, d *drawdomain.Draw) error {
	return f.err
}

func (f *failingDrawRepository) GetByPostID(ctx context.Context, postID post.DarkPostID) (*drawdomain.Draw, error) {
	return nil, f.err
}

func (f *failingDrawRepository) ListReady(ctx context.Context) ([]*drawdomain.Draw, error) {
	return nil, f.err
}

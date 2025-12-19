package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"backend/internal/adapter/http/handler"
	"backend/internal/adapter/repository/memory"
	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/repository"
	drawusecase "backend/internal/usecase/draw"
)

// Container は API で使用する依存を保持する。
type Container struct {
	Infra              *Infra
	DrawFortuneUsecase *drawusecase.FortuneUsecase
	DrawHandler        *handler.DrawHandler
}

// NewContainer は依存を初期化して返す。
func NewContainer(ctx context.Context) (*Container, error) {
	infra, err := NewInfra(ctx)
	if err != nil {
		return nil, fmt.Errorf("init infra: %w", err)
	}

	repo, err := provideDrawRepository()
	if err != nil {
		return nil, fmt.Errorf("provide draw repository: %w", err)
	}

	usecase := drawusecase.NewFortuneUsecase(repo)
	drawHandler := handler.NewDrawHandler(usecase)

	return &Container{
		Infra:              infra,
		DrawFortuneUsecase: usecase,
		DrawHandler:        drawHandler,
	}, nil
}

// Close は保持している外部リソースをクローズする。
func (c *Container) Close() error {
	if c == nil || c.Infra == nil {
		return nil
	}
	return c.Infra.Close()
}

func provideDrawRepository() (repository.DrawRepository, error) {
	mode := os.Getenv("DRAW_REPOSITORY_MODE")
	if mode == "error" {
		return newFailingDrawRepository(), nil
	}

	repo := memory.NewInMemoryDrawRepository()
	if err := seedDraws(repo, mode); err != nil {
		return nil, err
	}
	return repo, nil
}

func seedDraws(repo repository.DrawRepository, mode string) error {
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
		if err := repo.Create(context.Background(), draw); err != nil && err != repository.ErrDrawAlreadyExists {
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

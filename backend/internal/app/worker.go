package app

import (
	"context"
	"fmt"

	"backend/internal/adapter/llm/gemini"
	queueMemory "backend/internal/adapter/queue/memory"
	repoMemory "backend/internal/adapter/repository/memory"
	"backend/internal/config"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
	"backend/internal/usecase/worker"
)

// WorkerContainer はワーカーで使用する依存を保持する。
type WorkerContainer struct {
	Infra                *Infra
	PostRepo             repository.PostRepository
	JobQueue             queue.JobQueue
	Formatter            llm.Formatter
	FormatPendingUsecase *worker.FormatPendingUsecase
	closeFormatter       func() error
}

// NewWorkerContainer はワーカーの依存を初期化して返す。
func NewWorkerContainer(ctx context.Context) (*WorkerContainer, error) {
	infra, err := NewInfra(ctx)
	if err != nil {
		return nil, fmt.Errorf("init infra: %w", err)
	}

	postRepo := repoMemory.NewInMemoryPostRepository()
	initialID, err := seedPosts(postRepo)
	if err != nil {
		return nil, fmt.Errorf("seed posts: %w", err)
	}

	jobQueue := queueMemory.NewInMemoryJobQueue(10)

	geminiCfg, err := config.LoadGeminiConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("load gemini config: %w", err)
	}
	formatter, err := gemini.NewFormatter(ctx, geminiCfg.APIKey, geminiCfg.Model)
	if err != nil {
		return nil, fmt.Errorf("init formatter: %w", err)
	}

	if initialID != "" {
		_ = jobQueue.EnqueueFormat(ctx, initialID)
	}

	usecase := worker.NewFormatPendingUsecase(postRepo, formatter, jobQueue)

	return &WorkerContainer{
		Infra:                infra,
		PostRepo:             postRepo,
		JobQueue:             jobQueue,
		Formatter:            formatter,
		FormatPendingUsecase: usecase,
		closeFormatter:       formatter.Close,
	}, nil
}

// Close は保持している外部リソースをクローズする。
func (c *WorkerContainer) Close() error {
	if c == nil {
		return nil
	}
	if c.closeFormatter != nil {
		_ = c.closeFormatter()
	}
	if q, ok := c.JobQueue.(*queueMemory.InMemoryJobQueue); ok && q != nil {
		q.Close()
	}
	if c.Infra != nil {
		return c.Infra.Close()
	}
	return nil
}

func seedPosts(repo repository.PostRepository) (post.DarkPostID, error) {
	sample, err := post.New("post-local", "審査待ちの投稿です")
	if err != nil {
		return "", err
	}
	if err := repo.Create(context.Background(), sample); err != nil {
		return "", err
	}
	return sample.ID(), nil
}

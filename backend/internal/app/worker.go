package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"backend/internal/adapter/llm/gemini"
	queueMemory "backend/internal/adapter/queue/memory"
	repoFirestore "backend/internal/adapter/repository/firestore"
	repoMemory "backend/internal/adapter/repository/memory"
	"backend/internal/config"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
	"backend/internal/usecase/worker"
)

// ワーカーで使う依存をまとめた器。
type WorkerContainer struct {
	Infra                *Infra
	PostRepo             repository.PostRepository
	JobQueue             queue.JobQueue
	Formatter            llm.Formatter
	FormatPendingUsecase *worker.FormatPendingUsecase
	closeFormatter       func() error
}

/**
 * ワーカー稼働に必要なインフラ、LLM、キューなどを整えて返す。
 */
func NewWorkerContainer(ctx context.Context) (*WorkerContainer, error) {
	infra, err := NewInfra(ctx)
	if err != nil {
		return nil, fmt.Errorf("init infra: %w", err)
	}

	postRepo, seedLocal, err := newPostRepository(ctx, infra)
	if err != nil {
		return nil, err
	}
	var initialID post.DarkPostID
	if seedLocal {
		initialID, err = seedPosts(ctx, postRepo)
		if err != nil {
			return nil, fmt.Errorf("seed posts: %w", err)
		}
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

	// サンプル投稿があれば起動直後に処理させる
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

/**
 * 生成時に開いたリソースを順に閉じる。
 */
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

/**
 * メモリリポジトリへ見本投稿を入れ、整形対象 ID を返す。
 */
func seedPosts(ctx context.Context, repo repository.PostRepository) (post.DarkPostID, error) {
	sample, err := post.New("post-local", "審査待ちの投稿です")
	if err != nil {
		return "", err
	}
	if err := repo.Create(ctx, sample); err != nil {
		return "", err
	}
	return sample.ID(), nil
}

/**
 * 環境変数 WORKER_POST_REPOSITORY に応じてメモリ or Firestore のリポジトリを返す。
 */
func newPostRepository(ctx context.Context, infra *Infra) (repository.PostRepository, bool, error) {
	kind := strings.TrimSpace(os.Getenv("WORKER_POST_REPOSITORY"))
	switch strings.ToLower(kind) {
	case "firestore":
		client := infra.Firestore()
		if client == nil {
			return nil, false, fmt.Errorf("firestore post repository requested but firestore client is unavailable")
		}
		repo, err := repoFirestore.NewPostRepository(client)
		if err != nil {
			return nil, false, fmt.Errorf("new firestore post repository: %w", err)
		}
		return repo, false, nil
	default:
		return repoMemory.NewInMemoryPostRepository(), true, nil
	}
}

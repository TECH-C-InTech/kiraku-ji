package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"backend/internal/adapter/llm/gemini"
	openaiFormatter "backend/internal/adapter/llm/openai"
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
	closeInfra           func() error
}

var formatterCtor = gemini.NewFormatter

// OpenAI 用の整形器を作り、後片付け手順もあわせて返す
var openaiFormatterFactory = func(apiKey, model, baseURL string) (llm.Formatter, func() error, error) {
	formatter, err := openaiFormatter.NewFormatter(apiKey, model, baseURL)
	if err != nil {
		return nil, nil, err
	}
	return formatter, formatter.Close, nil
}

// 環境変数 LLM_PROVIDER に応じて利用する整形器を切り替える
var formatterFactory = func(ctx context.Context) (llm.Formatter, func() error, error) {
	switch config.LoadLLMProvider() {
	case "gemini":
		return newGeminiFormatter(ctx)
	case "openai":
		fallthrough
	default:
		return newOpenAIFormatter()
	}
}
var postRepositoryFactory = newPostRepository
var seedPostsFunc = seedPosts
var infraFactory = NewInfra
var samplePostFactory = func() (*post.Post, error) {
	return post.New("post-local", "審査待ちの投稿です")
}

/**
 * ワーカー稼働に必要なインフラ、LLM、キューなどを整えて返す。
 */
func NewWorkerContainer(ctx context.Context) (*WorkerContainer, error) {
	infra, err := infraFactory(ctx)
	if err != nil {
		return nil, fmt.Errorf("init infra: %w", err)
	}

	postRepo, seedLocal, err := postRepositoryFactory(ctx, infra)
	if err != nil {
		return nil, err
	}
	var initialID post.DarkPostID
	// メモリリポジトリ利用時のみサンプル投稿を投入しておく
	if seedLocal {
		initialID, err = seedPostsFunc(ctx, postRepo)
		if err != nil {
			return nil, fmt.Errorf("seed posts: %w", err)
		}
	}

	// ジョブキューは JOB_QUEUE_BACKEND で選択し、メモリ時は起動直後 enqueue のため seed フラグを受け取る
	jobQueue, seedMemoryQueue, err := jobQueueFactory(infra)
	if err != nil {
		return nil, fmt.Errorf("init job queue: %w", err)
	}

	// どの LLM プロバイダを使うかは formatterFactory が環境変数から判断する
	formatter, closeFormatter, err := formatterFactory(ctx)
	if err != nil {
		return nil, fmt.Errorf("init formatter: %w", err)
	}

	// サンプル投稿があれば起動直後に処理させる
	if initialID != "" && seedMemoryQueue {
		if err := jobQueue.EnqueueFormat(ctx, initialID); err != nil {
			log.Printf("seed enqueue failed: %v", err)
		}
	}

	usecase := worker.NewFormatPendingUsecase(postRepo, formatter, jobQueue)

	container := &WorkerContainer{
		Infra:                infra,
		PostRepo:             postRepo,
		JobQueue:             jobQueue,
		Formatter:            formatter,
		FormatPendingUsecase: usecase,
		closeFormatter:       closeFormatter,
	}
	if infra != nil {
		container.closeInfra = infra.Close
	}
	return container, nil
}

/**
 * 生成時に開いたリソースを順に閉じる。
 */
func (c *WorkerContainer) Close() error {
	if c == nil {
		return nil
	}
	var retErr error
	retErr = mergeCloseError(retErr, "formatter", c.closeFormatter)
	if c.JobQueue != nil {
		retErr = mergeCloseError(retErr, "job queue", c.JobQueue.Close)
	}
	return mergeCloseError(retErr, "infra", c.closeInfra)
}

/**
 * メモリリポジトリへ見本投稿を入れ、整形対象 ID を返す。
 */
func seedPosts(ctx context.Context, repo repository.PostRepository) (post.DarkPostID, error) {
	sample, err := samplePostFactory()
	if err != nil {
		return "", err
	}
	if err := repo.Create(ctx, sample); err != nil {
		return "", err
	}
	return sample.ID(), nil
}

/**
 * 閉じ処理を順番に呼び出し、最初に失敗したものを覚えて返す。
 */
func mergeCloseError(current error, label string, fn func() error) error {
	if fn == nil {
		// 閉じる対象がなければそのまま返す
		return current
	}
	// 後片付けの失敗はログに残しつつ先頭エラーを優先
	if err := fn(); err != nil {
		log.Printf("%s close error: %v", label, err)
		if current == nil {
			return err
		}
	}
	return current
}

/**
 * 環境変数から Gemini の鍵とモデルを読み込み、整形器とクローズ関数を返す。
 */
func newGeminiFormatter(ctx context.Context) (llm.Formatter, func() error, error) {
	// 鍵とモデル指定に不足がないかを先に確かめる
	cfg, err := config.LoadGeminiConfigFromEnv()
	if err != nil {
		return nil, nil, fmt.Errorf("load gemini config: %w", err)
	}
	// 構築済みクライアントを整形器として扱い、Close をそのまま返す
	formatter, err := formatterCtor(ctx, cfg.APIKey, cfg.Model)
	if err != nil {
		return nil, nil, fmt.Errorf("new gemini formatter: %w", err)
	}
	return formatter, formatter.Close, nil
}

/**
 * OpenAI 用の設定を取り込み、API クライアントを包んだ整形器を作る。
 */
func newOpenAIFormatter() (llm.Formatter, func() error, error) {
	// OpenAI 側の鍵やモデル、任意 BaseURL を取得
	cfg, err := config.LoadOpenAIConfigFromEnv()
	if err != nil {
		return nil, nil, fmt.Errorf("load openai config: %w", err)
	}
	// SDK から生成した整形器とクローズ処理を返す
	formatter, closeFn, err := openaiFormatterFactory(cfg.APIKey, cfg.Model, cfg.BaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("new openai formatter: %w", err)
	}
	return formatter, closeFn, nil
}

/**
 * 環境変数 WORKER_POST_REPOSITORY に応じてメモリ or Firestore のリポジトリを返す。
 */
func newPostRepository(ctx context.Context, infra *Infra) (repository.PostRepository, bool, error) {
	kind := strings.TrimSpace(os.Getenv("WORKER_POST_REPOSITORY"))
	switch strings.ToLower(kind) {
	case "firestore":
		repo, err := repoFirestore.NewPostRepository(infra.Firestore())
		if err != nil {
			return nil, false, fmt.Errorf("new firestore post repository: %w", err)
		}
		return repo, false, nil
	default:
		return repoMemory.NewInMemoryPostRepository(), true, nil
	}
}

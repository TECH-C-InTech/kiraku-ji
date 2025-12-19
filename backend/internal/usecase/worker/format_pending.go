package worker

import (
	"context"

	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
)

// FormatPendingUsecase は pending の闇投稿を整形するユースケース。
type FormatPendingUsecase struct {
	postRepo repository.PostRepository
	llm      llm.Formatter
	jobQueue queue.JobQueue
}

// NewFormatPendingUsecase は FormatPendingUsecase を初期化する。
func NewFormatPendingUsecase(
	postRepo repository.PostRepository,
	llmFormatter llm.Formatter,
	jobQueue queue.JobQueue,
) *FormatPendingUsecase {
	return &FormatPendingUsecase{
		postRepo: postRepo,
		llm:      llmFormatter,
		jobQueue: jobQueue,
	}
}

// Execute は与えられた闇投稿 ID を整形する。
func (u *FormatPendingUsecase) Execute(ctx context.Context, postID string) error {
	return nil
}

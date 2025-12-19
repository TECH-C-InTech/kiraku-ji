package worker

import (
	"context"
	"errors"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
)

var (
	// ErrEmptyPostID は入力 ID が空の場合に返される。
	ErrEmptyPostID = errors.New("format_pending: post id is empty")
	// ErrPostNotPending は pending 以外の投稿を整形しようとした場合に返される。
	ErrPostNotPending = errors.New("format_pending: post is not pending")
	// ErrNilUsecase は依存が未初期化の場合に返される。
	ErrNilUsecase = errors.New("format_pending: usecase is nil")
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

// Execute は与えられた闇投稿 ID を整形し、ready 状態へ遷移させる。
func (u *FormatPendingUsecase) Execute(ctx context.Context, postID string) error {
	if u == nil {
		return ErrNilUsecase
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if postID == "" {
		return ErrEmptyPostID
	}

	p, err := u.postRepo.Get(ctx, post.DarkPostID(postID))
	if err != nil {
		return err
	}
	if p.Status() != post.StatusPending {
		return ErrPostNotPending
	}

	formatResult, err := u.llm.Format(ctx, &llm.FormatRequest{
		DarkPostID:  p.ID(),
		DarkContent: p.Content(),
	})
	if err != nil {
		return err
	}

	validated, err := u.llm.Validate(ctx, formatResult)
	if err != nil {
		return err
	}

	if validated.Status != drawdomain.StatusVerified {
		return nil
	}

	if err := p.MarkReady(); err != nil {
		return err
	}

	return u.postRepo.Update(ctx, p)
}

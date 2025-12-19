package worker

import (
	"context"
	"errors"
	"fmt"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
)

var (
	ErrEmptyPostID          = errors.New("format_pending: 投稿 ID が指定されていません")
	ErrPostNotPending       = errors.New("format_pending: 整形待ちの投稿ではありません")
	ErrPostNotFound         = errors.New("format_pending: 投稿が存在しません")
	ErrFormatterUnavailable = errors.New("format_pending: 整形サービスに接続できません")
	ErrContentRejected      = errors.New("format_pending: 投稿内容が拒否されました")
	ErrNilUsecase           = errors.New("format_pending: ユースケースが初期化されていません")
	ErrNilContext           = errors.New("format_pending: コンテキストが指定されていません")
)

// 整形待ち投稿の整形から公開準備までを担う。
type FormatPendingUsecase struct {
	postRepo repository.PostRepository
	llm      llm.Formatter
	jobQueue queue.JobQueue // TODO: 再整形の再キュー処理で利用予定
}

// 依存をまとめて整形用ユースケースを組み立てる。
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

// LLM で整えて検証を通過した投稿を公開待ちに進める。
// 投稿欠如、LLM 停止など例外もエラーとして伝える。
func (u *FormatPendingUsecase) Execute(ctx context.Context, postID string) error {
	if u == nil {
		return ErrNilUsecase
	}
	if ctx == nil {
		return ErrNilContext
	}
	if postID == "" {
		return ErrEmptyPostID
	}

	p, err := u.postRepo.Get(ctx, post.DarkPostID(postID))
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return ErrPostNotFound
		}
		return err
	}

	formatResult, err := u.llm.Format(ctx, &llm.FormatRequest{
		DarkPostID:  p.ID(),
		DarkContent: p.Content(),
	})
	if err != nil {
		if errors.Is(err, llm.ErrFormatterUnavailable) {
			return ErrFormatterUnavailable
		}
		return err
	}

	validated, err := u.llm.Validate(ctx, formatResult)
	if err != nil {
		if errors.Is(err, llm.ErrContentRejected) {
			return ErrContentRejected
		}
		return err
	}

	// 検証で公開不可となった場合はここで終了
	if validated.Status != drawdomain.StatusVerified {
		return nil
	}

	// 公開待ちへの状態遷移に失敗した場合は元エラーも保持しつつ整形待ちではないとみなす
	if err := p.MarkReady(); err != nil {
		return fmt.Errorf("%w: %v", ErrPostNotPending, err)
	}

	return u.postRepo.Update(ctx, p)
}

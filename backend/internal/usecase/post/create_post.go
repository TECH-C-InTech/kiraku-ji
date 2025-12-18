package post

import (
	"context"
	"errors"

	"backend/internal/domain/post"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
)

var (
	// ErrNilInput はユースケースに nil 入力が渡された際に返される。
	ErrNilInput = errors.New("create_post: input is nil")
)

// 闇投稿作成の入力値
type CreatePostInput struct {
	DarkPostID string
	Content    string
}

// 闇投稿作成後に呼び出し側へ返す値
type CreatePostOutput struct {
	DarkPostID string
}

/**
 * 闇投稿作成のユースケース
 * postRepo: 投稿リポジトリ
 * jobQueue: 整形ジョブキュー
 */
type CreatePostUsecase struct {
	postRepo repository.PostRepository
	jobQueue queue.JobQueue
}

/**
 * ユースケース毎に初期化
 */
func NewCreatePostUsecase(postRepo repository.PostRepository, jobQueue queue.JobQueue) *CreatePostUsecase {
	return &CreatePostUsecase{
		postRepo: postRepo,
		jobQueue: jobQueue,
	}
}

/**
 * 闇投稿作成の実行
 */
func (u *CreatePostUsecase) Execute(ctx context.Context, in *CreatePostInput) (*CreatePostOutput, error) {
	if in == nil {
		return nil, ErrNilInput
	}

	p, err := post.New(post.DarkPostID(in.DarkPostID), post.DarkContent(in.Content))
	if err != nil {
		return nil, err
	}

	if err := u.postRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	if err := u.jobQueue.EnqueueFormat(ctx, p.ID()); err != nil {
		return nil, err
	}

	return &CreatePostOutput{DarkPostID: string(p.ID())}, nil
}

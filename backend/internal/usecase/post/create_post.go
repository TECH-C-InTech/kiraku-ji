package post

import (
	"context"

	"backend/internal/port/queue"
	"backend/internal/port/repository"
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
	// TODO: 実装（post.New -> PostRepository.Create -> JobQueue.EnqueueFormat）
	return nil, nil
}

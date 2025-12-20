package testutil

import (
	"context"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
)

// テストで投稿取得を差し替えるための簡易リポジトリ。
type StubPostRepository struct {
	Store     map[post.DarkPostID]*post.Post
	GetErr    error
	UpdateErr error
	Updated   *post.Post
}

/**
 * 投稿を 1 件だけ保持するスタブを返す。
 */
func NewStubPostRepository(p *post.Post) *StubPostRepository {
	store := make(map[post.DarkPostID]*post.Post)
	if p != nil {
		store[p.ID()] = p
	}
	return &StubPostRepository{Store: store}
}

/**
 * 作成呼び出しを無視する。
 */
func (r *StubPostRepository) Create(ctx context.Context, p *post.Post) error {
	return nil
}

/**
 * 保存された投稿か指定済みエラーを返す。
 */
func (r *StubPostRepository) Get(ctx context.Context, id post.DarkPostID) (*post.Post, error) {
	if r.GetErr != nil {
		return nil, r.GetErr
	}
	p, ok := r.Store[id]
	if !ok {
		return nil, repository.ErrPostNotFound
	}
	return p, nil
}

/**
 * ListReady は使用しないため nil を返す。
 */
func (r *StubPostRepository) ListReady(ctx context.Context, limit int) ([]*post.Post, error) {
	return nil, nil
}

/**
 * 更新内容を覚えて、必要ならエラーを返す。
 */
func (r *StubPostRepository) Update(ctx context.Context, p *post.Post) error {
	if r.UpdateErr != nil {
		return r.UpdateErr
	}
	r.Updated = p
	return nil
}

var _ repository.PostRepository = (*StubPostRepository)(nil)

// DrawRepository を埋めるだけの簡易スタブ。
type StubDrawRepository struct{}

/**
 * Create 呼び出しを無視する。
 */
func (StubDrawRepository) Create(ctx context.Context, d *drawdomain.Draw) error {
	return nil
}

/**
 * GetByPostID は既定で見つからない扱いにする。
 */
func (StubDrawRepository) GetByPostID(ctx context.Context, postID post.DarkPostID) (*drawdomain.Draw, error) {
	return nil, repository.ErrDrawNotFound
}

/**
 * ListReady は空を返す。
 */
func (StubDrawRepository) ListReady(ctx context.Context) ([]*drawdomain.Draw, error) {
	return nil, nil
}

var _ repository.DrawRepository = (*StubDrawRepository)(nil)

// 整形と検証の結果を切り替えられるテスト用スタブ。
type StubFormatter struct {
	FormatResult   *llm.FormatResult
	FormatErr      error
	ValidateResult *llm.FormatResult
	ValidateErr    error
}

/**
 * 設定された結果かエラーを返す。
 */
func (f *StubFormatter) Format(ctx context.Context, req *llm.FormatRequest) (*llm.FormatResult, error) {
	if f.FormatErr != nil {
		return nil, f.FormatErr
	}
	return f.FormatResult, nil
}

/**
 * 設定された結果かエラーを返す。
 */
func (f *StubFormatter) Validate(ctx context.Context, result *llm.FormatResult) (*llm.FormatResult, error) {
	if f.ValidateErr != nil {
		return nil, f.ValidateErr
	}
	return f.ValidateResult, nil
}

var _ llm.Formatter = (*StubFormatter)(nil)

// 依存を埋めるだけの空ジョブキュー。
type StubJobQueue struct{}

/**
 * 常に nil を返す。
 */
func (StubJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	return nil
}

/**
 * 閉鎖エラーを返す。
 */
func (StubJobQueue) DequeueFormat(ctx context.Context) (post.DarkPostID, error) {
	return "", queue.ErrQueueClosed
}

/**
 * 何もしない。
 */
func (StubJobQueue) Close() error {
	return nil
}

var _ queue.JobQueue = (*StubJobQueue)(nil)

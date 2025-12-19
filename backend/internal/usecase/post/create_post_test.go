package post

import (
	"context"
	"errors"
	"testing"

	"backend/internal/domain/post"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
)

func TestCreatePostUsecase_Execute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		input      *CreatePostInput
		setupRepo  func() *stubPostRepository
		setupQueue func() *stubJobQueue
		wantID     string
		wantErr    error
	}

	newUsecase := func(repo repository.PostRepository, q queue.JobQueue) *CreatePostUsecase {
		return NewCreatePostUsecase(repo, q)
	}

	cases := []testCase{
		{
			name:  "投稿保存とジョブ投入が成功する",
			input: &CreatePostInput{DarkPostID: "abc123", Content: "闇"},
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{
					createFunc: func(ctx context.Context, p *post.Post) error {
						if p.ID() != post.DarkPostID("abc123") {
							t.Fatalf("想定外の投稿ID: %s", p.ID())
						}
						if p.Content() != post.DarkContent("闇") {
							t.Fatalf("想定外の本文: %s", p.Content())
						}
						return nil
					},
				}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{
					enqueueFunc: func(ctx context.Context, id post.DarkPostID) error {
						if id != post.DarkPostID("abc123") {
							t.Fatalf("想定外のジョブ投入ID: %s", id)
						}
						return nil
					},
				}
			},
			wantID: "abc123",
		},
		{
			name:    "入力がnilなら ErrNilInput",
			input:   nil,
			wantErr: ErrNilInput,
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{}
			},
		},
		{
			name: "post.New のバリデーションエラーを返す",
			input: &CreatePostInput{
				DarkPostID: "abc123",
				Content:    "",
			},
			wantErr: post.ErrEmptyContent,
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{}
			},
		},
		{
			name: "リポジトリの重複エラーを変換する",
			input: &CreatePostInput{
				DarkPostID: "abc123",
				Content:    "闇",
			},
			wantErr: ErrPostAlreadyExists,
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{
					createFunc: func(ctx context.Context, p *post.Post) error {
						return repository.ErrPostAlreadyExists
					},
				}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{}
			},
		},
		{
			name: "リポジトリでの一般的なエラーはそのまま返す",
			input: &CreatePostInput{
				DarkPostID: "abc123",
				Content:    "闇",
			},
			wantErr: errors.New("リポジトリで異常が発生"),
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{
					createFunc: func(ctx context.Context, p *post.Post) error {
						return errors.New("リポジトリで異常が発生")
					},
				}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{}
			},
		},
		{
			name: "ジョブキューの重複エラーを変換する",
			input: &CreatePostInput{
				DarkPostID: "abc123",
				Content:    "闇",
			},
			wantErr: ErrJobAlreadyScheduled,
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{
					enqueueFunc: func(ctx context.Context, id post.DarkPostID) error {
						return queue.ErrJobAlreadyScheduled
					},
				}
			},
		},
		{
			name: "ジョブキューの一般的なエラーはそのまま返す",
			input: &CreatePostInput{
				DarkPostID: "abc123",
				Content:    "闇",
			},
			wantErr: errors.New("ジョブキューで異常が発生"),
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{
					enqueueFunc: func(ctx context.Context, id post.DarkPostID) error {
						return errors.New("ジョブキューで異常が発生")
					},
				}
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := tc.setupRepo()
			q := tc.setupQueue()

			uc := newUsecase(repo, q)
			got, err := uc.Execute(context.Background(), tc.input)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("エラー %v を期待したが nil", tc.wantErr)
				}
				if tc.wantErr.Error() != err.Error() {
					t.Fatalf("期待しないエラー: want %v, got %v", tc.wantErr, err)
				}

				if got != nil {
					t.Fatalf("出力は nil を期待したが %#v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("想定外のエラー: %v", err)
			}

			if got.DarkPostID != tc.wantID {
				t.Fatalf("返却IDが想定外: want %s, got %s", tc.wantID, got.DarkPostID)
			}
		})
	}
}

// stubPostRepository は PostRepository の簡易モック。
type stubPostRepository struct {
	createFunc func(context.Context, *post.Post) error
}

func (s *stubPostRepository) Create(ctx context.Context, p *post.Post) error {
	if s.createFunc != nil {
		return s.createFunc(ctx, p)
	}
	return nil
}

func (*stubPostRepository) Get(context.Context, post.DarkPostID) (*post.Post, error) {
	panic("not implemented")
}

func (*stubPostRepository) ListReady(context.Context, int) ([]*post.Post, error) {
	panic("not implemented")
}

func (*stubPostRepository) Update(context.Context, *post.Post) error {
	panic("not implemented")
}

// stubJobQueue は JobQueue の簡易モック。
type stubJobQueue struct {
	enqueueFunc func(context.Context, post.DarkPostID) error
}

func (s *stubJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	if s.enqueueFunc != nil {
		return s.enqueueFunc(ctx, id)
	}
	return nil
}

func (s *stubJobQueue) DequeueFormat(ctx context.Context) (post.DarkPostID, error) {
	return "", queue.ErrQueueClosed
}

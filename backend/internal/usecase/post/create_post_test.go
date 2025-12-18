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
			name:  "successfully creates post and enqueues job",
			input: &CreatePostInput{DarkPostID: "abc123", Content: "闇"},
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{
					createFunc: func(ctx context.Context, p *post.Post) error {
						if p.ID() != post.DarkPostID("abc123") {
							t.Fatalf("unexpected post id: %s", p.ID())
						}
						if p.Content() != post.DarkContent("闇") {
							t.Fatalf("unexpected content: %s", p.Content())
						}
						return nil
					},
				}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{
					enqueueFunc: func(ctx context.Context, id post.DarkPostID) error {
						if id != post.DarkPostID("abc123") {
							t.Fatalf("unexpected enqueue id: %s", id)
						}
						return nil
					},
				}
			},
			wantID: "abc123",
		},
		{
			name:    "nil input",
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
			name: "post validation error surfaces",
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
			name: "repository sentinel converted",
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
			name: "repository general error bubbles",
			input: &CreatePostInput{
				DarkPostID: "abc123",
				Content:    "闇",
			},
			wantErr: errors.New("repo boom"),
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{
					createFunc: func(ctx context.Context, p *post.Post) error {
						return errors.New("repo boom")
					},
				}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{}
			},
		},
		{
			name: "queue sentinel converted",
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
			name: "queue general error bubbles",
			input: &CreatePostInput{
				DarkPostID: "abc123",
				Content:    "闇",
			},
			wantErr: errors.New("queue boom"),
			setupRepo: func() *stubPostRepository {
				return &stubPostRepository{}
			},
			setupQueue: func() *stubJobQueue {
				return &stubJobQueue{
					enqueueFunc: func(ctx context.Context, id post.DarkPostID) error {
						return errors.New("queue boom")
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
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if tc.wantErr.Error() != err.Error() {
					t.Fatalf("unexpected error: want %v, got %v", tc.wantErr, err)
				}

				if got != nil {
					t.Fatalf("expected nil output, got %#v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.DarkPostID != tc.wantID {
				t.Fatalf("unexpected returned id: want %s, got %s", tc.wantID, got.DarkPostID)
			}
		})
	}
}

// stubPostRepository implements repository.PostRepository for tests.
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

// stubJobQueue implements queue.JobQueue for tests.
type stubJobQueue struct {
	enqueueFunc func(context.Context, post.DarkPostID) error
}

func (s *stubJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	if s.enqueueFunc != nil {
		return s.enqueueFunc(ctx, id)
	}
	return nil
}

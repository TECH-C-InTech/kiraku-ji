package worker

import (
	"context"
	"errors"
	"testing"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
)

type stubPostRepository struct {
	store     map[post.DarkPostID]*post.Post
	getErr    error
	updateErr error
	updated   *post.Post
}

var _ repository.PostRepository = (*stubPostRepository)(nil)

func newStubPostRepository(p *post.Post) *stubPostRepository {
	store := make(map[post.DarkPostID]*post.Post)
	if p != nil {
		store[p.ID()] = p
	}
	return &stubPostRepository{store: store}
}

func (r *stubPostRepository) Create(ctx context.Context, p *post.Post) error {
	return nil
}

func (r *stubPostRepository) Get(ctx context.Context, id post.DarkPostID) (*post.Post, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	p, ok := r.store[id]
	if !ok {
		return nil, repository.ErrPostNotFound
	}
	return p, nil
}

func (r *stubPostRepository) ListReady(ctx context.Context, limit int) ([]*post.Post, error) {
	return nil, nil
}

func (r *stubPostRepository) Update(ctx context.Context, p *post.Post) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updated = p
	return nil
}

type stubFormatter struct {
	formatResult   *llm.FormatResult
	formatErr      error
	validateResult *llm.FormatResult
	validateErr    error
}

var _ llm.Formatter = (*stubFormatter)(nil)

func (f *stubFormatter) Format(ctx context.Context, req *llm.FormatRequest) (*llm.FormatResult, error) {
	if f.formatErr != nil {
		return nil, f.formatErr
	}
	return f.formatResult, nil
}

func (f *stubFormatter) Validate(ctx context.Context, result *llm.FormatResult) (*llm.FormatResult, error) {
	if f.validateErr != nil {
		return nil, f.validateErr
	}
	return f.validateResult, nil
}

type stubJobQueue struct{}

func (stubJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	return nil
}

var _ queue.JobQueue = (*stubJobQueue)(nil)

func TestFormatPendingUsecase_Success(t *testing.T) {
	p, err := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}
	repo := newStubPostRepository(p)
	formatter := &stubFormatter{
		formatResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusPending,
			FormattedContent: "formatted",
		},
		validateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusVerified,
			FormattedContent: "formatted",
		},
	}
	usecase := NewFormatPendingUsecase(repo, formatter, stubJobQueue{})

	if err := usecase.Execute(context.Background(), "post-1"); err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if repo.updated == nil {
		t.Fatalf("expected update to be called")
	}
	if repo.updated.Status() != post.StatusReady {
		t.Fatalf("expected post to be marked ready, status=%s", repo.updated.Status())
	}
}

func TestFormatPendingUsecase_PostNotFound(t *testing.T) {
	repo := newStubPostRepository(nil)
	usecase := NewFormatPendingUsecase(repo, &stubFormatter{}, stubJobQueue{})

	err := usecase.Execute(context.Background(), "unknown")
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestFormatPendingUsecase_FormatterUnavailable(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := newStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, &stubFormatter{
		formatErr: llm.ErrFormatterUnavailable,
	}, stubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, ErrFormatterUnavailable) {
		t.Fatalf("expected ErrFormatterUnavailable, got %v", err)
	}
}

func TestFormatPendingUsecase_ContentRejected(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := newStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, &stubFormatter{
		formatResult: &llm.FormatResult{DarkPostID: p.ID()},
		validateErr:  llm.ErrContentRejected,
	}, stubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, ErrContentRejected) {
		t.Fatalf("expected ErrContentRejected, got %v", err)
	}
	if repo.updated != nil {
		t.Fatalf("post should not be updated on rejection")
	}
}

func TestFormatPendingUsecase_UpdateFailed(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := newStubPostRepository(p)
	repo.updateErr = errors.New("update failed")
	usecase := NewFormatPendingUsecase(repo, &stubFormatter{
		formatResult: &llm.FormatResult{DarkPostID: p.ID()},
		validateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusVerified,
			FormattedContent: "formatted",
		},
	}, stubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if err == nil || !errors.Is(err, repo.updateErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestFormatPendingUsecase_GetGenericError(t *testing.T) {
	repo := newStubPostRepository(nil)
	repo.getErr = errors.New("db down")
	usecase := NewFormatPendingUsecase(repo, &stubFormatter{}, stubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, repo.getErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestFormatPendingUsecase_FormatGenericError(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := newStubPostRepository(p)
	expectedErr := errors.New("format failed")
	usecase := NewFormatPendingUsecase(repo, &stubFormatter{
		formatErr: expectedErr,
	}, stubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected generic format error, got %v", err)
	}
}

func TestFormatPendingUsecase_ValidateGenericError(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := newStubPostRepository(p)
	expectedErr := errors.New("validate failed")
	usecase := NewFormatPendingUsecase(repo, &stubFormatter{
		formatResult: &llm.FormatResult{DarkPostID: p.ID()},
		validateErr:  expectedErr,
	}, stubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected generic validate error, got %v", err)
	}
}

func TestFormatPendingUsecase_PostNotPending(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	if err := p.MarkReady(); err != nil {
		t.Fatalf("failed to mark ready: %v", err)
	}
	repo := newStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, &stubFormatter{
		formatResult: &llm.FormatResult{DarkPostID: p.ID()},
		validateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusVerified,
			FormattedContent: "formatted",
		},
	}, stubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, ErrPostNotPending) {
		t.Fatalf("expected ErrPostNotPending, got %v", err)
	}
}

func TestFormatPendingUsecase_EmptyPostID(t *testing.T) {
	usecase := NewFormatPendingUsecase(newStubPostRepository(nil), &stubFormatter{}, stubJobQueue{})
	if err := usecase.Execute(context.Background(), ""); !errors.Is(err, ErrEmptyPostID) {
		t.Fatalf("expected ErrEmptyPostID, got %v", err)
	}
}

func TestFormatPendingUsecase_NilUsecase(t *testing.T) {
	var usecase *FormatPendingUsecase
	if err := usecase.Execute(context.Background(), "post-1"); !errors.Is(err, ErrNilUsecase) {
		t.Fatalf("expected ErrNilUsecase, got %v", err)
	}
}

func TestFormatPendingUsecase_ValidationNotVerified(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := newStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, &stubFormatter{
		formatResult: &llm.FormatResult{DarkPostID: p.ID()},
		validateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusRejected,
			FormattedContent: "formatted",
		},
	}, stubJobQueue{})

	if err := usecase.Execute(context.Background(), "post-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updated != nil {
		t.Fatalf("post should not be updated when not verified")
	}
}

func TestFormatPendingUsecase_NilContext(t *testing.T) {
	usecase := NewFormatPendingUsecase(newStubPostRepository(nil), &stubFormatter{}, stubJobQueue{})

	var nilCtx context.Context
	if err := usecase.Execute(nilCtx, "post-1"); !errors.Is(err, ErrNilContext) {
		t.Fatalf("expected ErrNilContext, got %v", err)
	}
}

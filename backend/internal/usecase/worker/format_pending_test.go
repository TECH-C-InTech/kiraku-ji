package worker

import (
	"context"
	"errors"
	"testing"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/usecase/worker/testutil"
)

func TestFormatPendingUsecase_Success(t *testing.T) {
	p, err := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}
	repo := testutil.NewStubPostRepository(p)
	formatter := &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusPending,
			FormattedContent: "formatted",
		},
		ValidateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusVerified,
			FormattedContent: "formatted",
		},
	}
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, formatter, testutil.StubJobQueue{})

	if err := usecase.Execute(context.Background(), "post-1"); err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if repo.Updated == nil {
		t.Fatalf("expected update to be called")
	}
	if repo.Updated.Status() != post.StatusReady {
		t.Fatalf("expected post to be marked ready, status=%s", repo.Updated.Status())
	}
}

func TestFormatPendingUsecase_PostNotFound(t *testing.T) {
	repo := testutil.NewStubPostRepository(nil)
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, &testutil.StubFormatter{}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "unknown")
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestFormatPendingUsecase_GetGenericError(t *testing.T) {
	repo := testutil.NewStubPostRepository(nil)
	repo.GetErr = errors.New("get failed")
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, &testutil.StubFormatter{}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, repo.GetErr) {
		t.Fatalf("expected generic get error, got %v", err)
	}
}

func TestFormatPendingUsecase_FormatterUnavailable(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, &testutil.StubFormatter{
		FormatErr: llm.ErrFormatterUnavailable,
	}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, ErrFormatterUnavailable) {
		t.Fatalf("expected ErrFormatterUnavailable, got %v", err)
	}
}

func TestFormatPendingUsecase_ContentRejected(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{DarkPostID: p.ID()},
		ValidateErr:  llm.ErrContentRejected,
	}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, ErrContentRejected) {
		t.Fatalf("expected ErrContentRejected, got %v", err)
	}
	if repo.Updated != nil {
		t.Fatalf("post should not be updated on rejection")
	}
}

func TestFormatPendingUsecase_UpdateFailed(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	repo.UpdateErr = errors.New("update failed")
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{DarkPostID: p.ID()},
		ValidateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusVerified,
			FormattedContent: "formatted",
		},
	}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if err == nil || !errors.Is(err, repo.UpdateErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestFormatPendingUsecase_PostNotPending(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	if err := p.MarkReady(); err != nil {
		t.Fatalf("failed to mark ready: %v", err)
	}
	repo := testutil.NewStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{DarkPostID: p.ID()},
		ValidateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusVerified,
			FormattedContent: "formatted",
		},
	}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, ErrPostNotPending) {
		t.Fatalf("expected ErrPostNotPending, got %v", err)
	}
}

func TestFormatPendingUsecase_EmptyPostID(t *testing.T) {
	usecase := NewFormatPendingUsecase(testutil.NewStubPostRepository(nil), testutil.StubDrawRepository{}, &testutil.StubFormatter{}, testutil.StubJobQueue{})
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
	repo := testutil.NewStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{DarkPostID: p.ID()},
		ValidateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusRejected,
			FormattedContent: "formatted",
		},
	}, testutil.StubJobQueue{})

	if err := usecase.Execute(context.Background(), "post-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Updated != nil {
		t.Fatalf("post should not be updated when not verified")
	}
}

func TestFormatPendingUsecase_NilContext(t *testing.T) {
	usecase := NewFormatPendingUsecase(testutil.NewStubPostRepository(nil), testutil.StubDrawRepository{}, &testutil.StubFormatter{}, testutil.StubJobQueue{})

	var nilCtx context.Context
	if err := usecase.Execute(nilCtx, "post-1"); !errors.Is(err, ErrNilContext) {
		t.Fatalf("expected ErrNilContext, got %v", err)
	}
}

func TestFormatPendingUsecase_FormatGenericError(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	expectedErr := errors.New("format failed")
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, &testutil.StubFormatter{
		FormatErr: expectedErr,
	}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected generic format error, got %v", err)
	}
}

func TestFormatPendingUsecase_ValidateGenericError(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	expectedErr := errors.New("validate failed")
	usecase := NewFormatPendingUsecase(repo, testutil.StubDrawRepository{}, &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{DarkPostID: p.ID()},
		ValidateErr:  expectedErr,
	}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected generic validate error, got %v", err)
	}
}

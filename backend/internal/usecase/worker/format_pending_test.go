package worker

import (
	"context"
	"errors"
	"strings"
	"testing"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/usecase/worker/testutil"
)

func TestFormatPendingUsecase_Success(t *testing.T) {
	p, err := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}
	repo := testutil.NewStubPostRepository(p)
	drawRepo := &testutil.StubDrawRepository{}
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
	usecase := NewFormatPendingUsecase(repo, drawRepo, formatter, testutil.StubJobQueue{})

	if err := usecase.Execute(context.Background(), "post-1"); err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if repo.Updated == nil {
		t.Fatalf("expected update to be called")
	}
	if repo.Updated.Status() != post.StatusReady {
		t.Fatalf("expected post to be marked ready, status=%s", repo.Updated.Status())
	}
	if len(drawRepo.Created) != 1 {
		t.Fatalf("expected draw to be created once, got %d", len(drawRepo.Created))
	}
	created := drawRepo.Created[0]
	if created.PostID() != p.ID() {
		t.Fatalf("unexpected draw post id: %s", created.PostID())
	}
	if created.Result() != drawdomain.FormattedContent("formatted") {
		t.Fatalf("unexpected draw result: %s", created.Result())
	}
	if created.Status() != drawdomain.StatusVerified {
		t.Fatalf("expected verified draw, got %s", created.Status())
	}
}

func TestFormatPendingUsecase_PostNotFound(t *testing.T) {
	repo := testutil.NewStubPostRepository(nil)
	usecase := NewFormatPendingUsecase(repo, &testutil.StubDrawRepository{}, &testutil.StubFormatter{}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "unknown")
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestFormatPendingUsecase_GetGenericError(t *testing.T) {
	repo := testutil.NewStubPostRepository(nil)
	repo.GetErr = errors.New("get failed")
	usecase := NewFormatPendingUsecase(repo, &testutil.StubDrawRepository{}, &testutil.StubFormatter{}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, repo.GetErr) {
		t.Fatalf("expected generic get error, got %v", err)
	}
}

func TestFormatPendingUsecase_FormatterUnavailable(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, &testutil.StubDrawRepository{}, &testutil.StubFormatter{
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
	usecase := NewFormatPendingUsecase(repo, &testutil.StubDrawRepository{}, &testutil.StubFormatter{
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
	drawRepo := &testutil.StubDrawRepository{}
	usecase := NewFormatPendingUsecase(repo, drawRepo, &testutil.StubFormatter{
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
	if len(drawRepo.Created) != 1 {
		t.Fatalf("expected draw to be created before update")
	}
}

func TestFormatPendingUsecase_PostNotPending(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	if err := p.MarkReady(); err != nil {
		t.Fatalf("failed to mark ready: %v", err)
	}
	repo := testutil.NewStubPostRepository(p)
	usecase := NewFormatPendingUsecase(repo, &testutil.StubDrawRepository{}, &testutil.StubFormatter{
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
	usecase := NewFormatPendingUsecase(testutil.NewStubPostRepository(nil), &testutil.StubDrawRepository{}, &testutil.StubFormatter{}, testutil.StubJobQueue{})
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
	usecase := NewFormatPendingUsecase(repo, &testutil.StubDrawRepository{}, &testutil.StubFormatter{
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
	usecase := NewFormatPendingUsecase(testutil.NewStubPostRepository(nil), &testutil.StubDrawRepository{}, &testutil.StubFormatter{}, testutil.StubJobQueue{})

	var nilCtx context.Context
	if err := usecase.Execute(nilCtx, "post-1"); !errors.Is(err, ErrNilContext) {
		t.Fatalf("expected ErrNilContext, got %v", err)
	}
}

func TestFormatPendingUsecase_FormatGenericError(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	expectedErr := errors.New("format failed")
	usecase := NewFormatPendingUsecase(repo, &testutil.StubDrawRepository{}, &testutil.StubFormatter{
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
	usecase := NewFormatPendingUsecase(repo, &testutil.StubDrawRepository{}, &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{DarkPostID: p.ID()},
		ValidateErr:  expectedErr,
	}, testutil.StubJobQueue{})

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected generic validate error, got %v", err)
	}
}

func TestFormatPendingUsecase_DrawCreateFailed(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	drawRepo := &testutil.StubDrawRepository{
		CreateErr: errors.New("draw create failed"),
	}
	jobQueue := &recordingJobQueue{}
	usecase := NewFormatPendingUsecase(repo, drawRepo, &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{DarkPostID: p.ID()},
		ValidateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusVerified,
			FormattedContent: "formatted",
		},
	}, jobQueue)

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, ErrDrawCreationFailed) {
		t.Fatalf("expected ErrDrawCreationFailed, got %v", err)
	}
	if repo.Updated != nil {
		t.Fatalf("post should not be updated when draw creation fails")
	}
	if len(drawRepo.Created) != 0 {
		t.Fatalf("draw should not be recorded when create fails")
	}
	if len(jobQueue.enqueued) != 1 || jobQueue.enqueued[0] != p.ID() {
		t.Fatalf("expected post to be requeued once")
	}
}

func TestFormatPendingUsecase_DrawContentTrimmedAndLimited(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	drawRepo := &testutil.StubDrawRepository{}
	raw := drawdomain.FormattedContent("  \n" + strings.Repeat("運", maxDrawResultLength+5) + "  ")

	usecase := NewFormatPendingUsecase(repo, drawRepo, &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusPending,
			FormattedContent: raw,
		},
		ValidateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusVerified,
			FormattedContent: raw,
		},
	}, testutil.StubJobQueue{})

	if err := usecase.Execute(context.Background(), "post-1"); err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if len(drawRepo.Created) != 1 {
		t.Fatalf("expected draw creation")
	}
	resultRunes := []rune(string(drawRepo.Created[0].Result()))
	if len(resultRunes) != maxDrawResultLength {
		t.Fatalf("expected trimmed result length %d, got %d", maxDrawResultLength, len(resultRunes))
	}
	if resultRunes[0] != '運' || resultRunes[len(resultRunes)-1] != '運' {
		t.Fatalf("expected spaces and newlines to be trimmed")
	}
}

type recordingJobQueue struct {
	enqueued   []post.DarkPostID
	enqueueErr error
}

func (q *recordingJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	if q.enqueueErr != nil {
		return q.enqueueErr
	}
	q.enqueued = append(q.enqueued, id)
	return nil
}

func (*recordingJobQueue) DequeueFormat(ctx context.Context) (post.DarkPostID, error) {
	return "", queue.ErrQueueClosed
}

func (*recordingJobQueue) Close() error {
	return nil
}

func TestFormatPendingUsecase_DrawCreateFailed_RequeueError(t *testing.T) {
	p, _ := post.New(post.DarkPostID("post-1"), post.DarkContent("test"))
	repo := testutil.NewStubPostRepository(p)
	drawRepo := &testutil.StubDrawRepository{
		CreateErr: errors.New("draw create failed"),
	}
	jobQueue := &recordingJobQueue{
		enqueueErr: errors.New("queue down"),
	}
	usecase := NewFormatPendingUsecase(repo, drawRepo, &testutil.StubFormatter{
		FormatResult: &llm.FormatResult{DarkPostID: p.ID()},
		ValidateResult: &llm.FormatResult{
			DarkPostID:       p.ID(),
			Status:           drawdomain.StatusVerified,
			FormattedContent: "formatted",
		},
	}, jobQueue)

	err := usecase.Execute(context.Background(), "post-1")
	if !errors.Is(err, ErrRequeueFailed) {
		t.Fatalf("expected ErrRequeueFailed, got %v", err)
	}
	if repo.Updated != nil {
		t.Fatalf("post should not be updated when requeue fails")
	}
	if len(jobQueue.enqueued) != 0 {
		t.Fatalf("requeue should not record success when enqueue fails")
	}
}

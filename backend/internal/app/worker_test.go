package app

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	"backend/internal/adapter/llm/gemini"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
)

func TestNewWorkerContainer_UsesFirestoreRepository(t *testing.T) {
	setRequiredFirestoreEnv(t)

	stubFormatter := &stubFormatter{}
	origFormatterFactory := formatterFactory
	formatterFactory = func(ctx context.Context) (llm.Formatter, func() error, error) {
		return stubFormatter, stubFormatter.Close, nil
	}
	defer func() { formatterFactory = origFormatterFactory }()

	stubRepo := &workerStubPostRepository{}
	origRepoFactory := postRepositoryFactory
	postRepositoryFactory = func(ctx context.Context, infra *Infra) (repository.PostRepository, error) {
		return stubRepo, nil
	}
	defer func() { postRepositoryFactory = origRepoFactory }()

	origInfraFactory := infraFactory
	infraFactory = func(ctx context.Context) (*Infra, error) {
		return &Infra{}, nil
	}
	defer func() { infraFactory = origInfraFactory }()

	container, err := NewWorkerContainer(context.Background())
	if err != nil {
		t.Fatalf("NewWorkerContainer returned error: %v", err)
	}
	if container.PostRepo != stubRepo {
		t.Fatalf("expected stub repository to be used")
	}
	if err := container.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}
	if !stubFormatter.closed {
		t.Fatalf("formatter should be closable via container.Close")
	}
}

func TestFormatterFactory_UsesCtor(t *testing.T) {
	origCtor := formatterCtor
	stub := &gemini.Formatter{}
	formatterCtor = func(ctx context.Context, apiKey, model string, opts ...option.ClientOption) (*gemini.Formatter, error) {
		return stub, nil
	}
	defer func() { formatterCtor = origCtor }()

	t.Setenv("GEMINI_API_KEY", "key")
	t.Setenv("GEMINI_MODEL", "model")

	f, closer, err := newGeminiFormatter(context.Background())
	if err != nil {
		t.Fatalf("newGeminiFormatter returned error: %v", err)
	}
	if f != stub {
		t.Fatalf("expected stub formatter")
	}
	if closer == nil {
		t.Fatalf("expected close func")
	}
}
func TestNewWorkerContainer_PostRepoError(t *testing.T) {
	setRequiredFirestoreEnv(t)

	origInfra := infraFactory
	infraFactory = func(ctx context.Context) (*Infra, error) {
		return &Infra{}, nil
	}
	defer func() { infraFactory = origInfra }()

	origRepoFactory := postRepositoryFactory
	postRepositoryFactory = func(ctx context.Context, infra *Infra) (repository.PostRepository, error) {
		return nil, errors.New("repo factory error")
	}
	defer func() { postRepositoryFactory = origRepoFactory }()

	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when repository factory fails")
	}
}

func TestNewWorkerContainer_FormatterFactoryError(t *testing.T) {
	setRequiredFirestoreEnv(t)

	origInfra := infraFactory
	infraFactory = func(ctx context.Context) (*Infra, error) {
		return &Infra{}, nil
	}
	defer func() { infraFactory = origInfra }()

	origRepoFactory := postRepositoryFactory
	postRepositoryFactory = func(ctx context.Context, infra *Infra) (repository.PostRepository, error) {
		return &workerStubPostRepository{}, nil
	}
	defer func() { postRepositoryFactory = origRepoFactory }()

	origFormatterFactory := formatterFactory
	formatterFactory = func(ctx context.Context) (llm.Formatter, func() error, error) {
		return nil, nil, errors.New("formatter error")
	}
	defer func() { formatterFactory = origFormatterFactory }()

	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when formatter factory fails")
	}
}

func TestNewWorkerContainer_MissingGeminiConfig(t *testing.T) {
	setRequiredFirestoreEnv(t)
	t.Setenv("LLM_PROVIDER", "gemini")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GEMINI_MODEL", "")

	origInfra := infraFactory
	infraFactory = func(ctx context.Context) (*Infra, error) {
		return &Infra{}, nil
	}
	defer func() { infraFactory = origInfra }()

	origRepoFactory := postRepositoryFactory
	postRepositoryFactory = func(ctx context.Context, infra *Infra) (repository.PostRepository, error) {
		return &workerStubPostRepository{}, nil
	}
	defer func() { postRepositoryFactory = origRepoFactory }()

	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when gemini config is missing")
	}
}

func TestNewWorkerContainer_InfraError(t *testing.T) {
	setRequiredFirestoreEnv(t)

	origInfra := infraFactory
	infraFactory = func(ctx context.Context) (*Infra, error) {
		return nil, errors.New("infra error")
	}
	defer func() { infraFactory = origInfra }()

	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when infra initialization fails")
	}
}

func TestNewWorkerContainer_FirestoreEnvMissing(t *testing.T) {
	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when firestore env vars are missing")
	} else if !errors.Is(err, errWorkerFirestoreEnvMissing) {
		t.Fatalf("expected missing env error, got %v", err)
	}
}

func TestNewWorkerContainer_JobQueueFactoryError(t *testing.T) {
	setRequiredFirestoreEnv(t)

	origInfra := infraFactory
	infraFactory = func(ctx context.Context) (*Infra, error) {
		return &Infra{}, nil
	}
	defer func() { infraFactory = origInfra }()

	origRepoFactory := postRepositoryFactory
	postRepositoryFactory = func(ctx context.Context, infra *Infra) (repository.PostRepository, error) {
		return &workerStubPostRepository{}, nil
	}
	defer func() { postRepositoryFactory = origRepoFactory }()

	origJobQueueFactory := jobQueueFactory
	jobQueueFactory = func(infra *Infra) (queue.JobQueue, bool, error) {
		return nil, false, errors.New("job queue error")
	}
	defer func() { jobQueueFactory = origJobQueueFactory }()

	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when job queue factory fails")
	}
}

func TestNewPostRepository_FirestoreRequiresClient(t *testing.T) {
	if _, err := newPostRepository(context.Background(), &Infra{}); err == nil {
		t.Fatalf("expected error when firestore client is missing")
	}
}

func TestNewPostRepository_FirestoreSuccess(t *testing.T) {
	infra := &Infra{firestoreClient: &firestore.Client{}}
	if _, err := newPostRepository(context.Background(), infra); err != nil {
		t.Fatalf("expected firestore repo, got error: %v", err)
	}
}

func TestWorkerContainerClose_ReturnsFirstError(t *testing.T) {
	queueStub := &stubJobQueue{closeErr: errors.New("queue close")}
	formatter := &stubFormatter{closeErr: errors.New("formatter close")}
	infraErr := errors.New("infra close")

	container := &WorkerContainer{
		JobQueue:       queueStub,
		closeFormatter: formatter.Close,
		closeInfra: func() error {
			return infraErr
		},
	}

	err := container.Close()
	if !errors.Is(err, formatter.closeErr) {
		t.Fatalf("expected formatter close error, got %v", err)
	}
	if !formatter.closed {
		t.Fatalf("formatter close was not invoked")
	}
	if !queueStub.closed {
		t.Fatalf("queue close was not invoked")
	}
}

func TestWorkerContainerClose_Nil(t *testing.T) {
	var container *WorkerContainer
	if err := container.Close(); err != nil {
		t.Fatalf("expected nil error for nil receiver")
	}
}

func TestMergeCloseError(t *testing.T) {
	if err := mergeCloseError(nil, "noop", nil); err != nil {
		t.Fatalf("expected nil when fn is nil")
	}

	baseErr := errors.New("base")
	nextErr := errors.New("next")
	result := mergeCloseError(baseErr, "label", func() error { return nextErr })
	if !errors.Is(result, baseErr) {
		t.Fatalf("expected base error to remain")
	}
	result = mergeCloseError(nil, "label", func() error { return nextErr })
	if !errors.Is(result, nextErr) {
		t.Fatalf("expected new error when base is nil")
	}
}

func TestNewOpenAIFormatter_Success(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_MODEL", "")
	t.Setenv("OPENAI_BASE_URL", "")

	stub := &stubFormatter{}
	origFactory := openaiFormatterFactory
	openaiFormatterFactory = func(apiKey, model, baseURL string) (llm.Formatter, func() error, error) {
		if apiKey != "test-key" {
			t.Fatalf("unexpected api key: %s", apiKey)
		}
		if model != "gpt-4o-mini" {
			t.Fatalf("expected default model, got %s", model)
		}
		return stub, stub.Close, nil
	}
	defer func() { openaiFormatterFactory = origFactory }()

	formatter, closer, err := newOpenAIFormatter()
	if err != nil {
		t.Fatalf("newOpenAIFormatter returned error: %v", err)
	}
	if formatter != stub {
		t.Fatalf("expected stub formatter")
	}
	if closer == nil {
		t.Fatalf("expected close function")
	}
}

func TestNewOpenAIFormatter_MissingConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	if _, _, err := newOpenAIFormatter(); err == nil {
		t.Fatalf("expected error when OPENAI_API_KEY is missing")
	}
}

func setRequiredFirestoreEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/service-account.json")
}

type stubFormatter struct {
	closeErr error
	closed   bool
}

func (s *stubFormatter) Format(ctx context.Context, req *llm.FormatRequest) (*llm.FormatResult, error) {
	return nil, nil
}

func (s *stubFormatter) Validate(ctx context.Context, result *llm.FormatResult) (*llm.FormatResult, error) {
	return result, nil
}

func (s *stubFormatter) Close() error {
	s.closed = true
	return s.closeErr
}

type stubJobQueue struct {
	closeErr error
	closed   bool
}

func (s *stubJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	return nil
}

func (s *stubJobQueue) DequeueFormat(ctx context.Context) (post.DarkPostID, error) {
	return "", nil
}

func (s *stubJobQueue) Close() error {
	s.closed = true
	if s.closeErr == nil {
		s.closeErr = queue.ErrQueueClosed
	}
	return s.closeErr
}

var _ llm.Formatter = (*stubFormatter)(nil)
var _ queue.JobQueue = (*stubJobQueue)(nil)
var _ repository.PostRepository = (*workerStubPostRepository)(nil)

type workerStubPostRepository struct{}

func (workerStubPostRepository) Create(ctx context.Context, p *post.Post) error {
	return nil
}

func (workerStubPostRepository) Get(ctx context.Context, id post.DarkPostID) (*post.Post, error) {
	return nil, repository.ErrPostNotFound
}

func (workerStubPostRepository) ListReady(ctx context.Context, limit int) ([]*post.Post, error) {
	return nil, nil
}

func (workerStubPostRepository) Update(ctx context.Context, p *post.Post) error {
	return repository.ErrPostNotFound
}

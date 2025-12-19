package app

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	"backend/internal/adapter/llm/gemini"
	repoMemory "backend/internal/adapter/repository/memory"
	"backend/internal/domain/post"
	"backend/internal/port/llm"
	"backend/internal/port/queue"
	"backend/internal/port/repository"
)

func TestNewWorkerContainer_MemorySeedsJob(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "dummy")
	t.Setenv("GEMINI_MODEL", "dummy-model")
	t.Setenv("WORKER_POST_REPOSITORY", "")

	stub := &stubFormatter{}
	origFactory := formatterFactory
	formatterFactory = func(ctx context.Context) (llm.Formatter, func() error, error) {
		return stub, stub.Close, nil
	}
	defer func() { formatterFactory = origFactory }()

	container, err := NewWorkerContainer(context.Background())
	if err != nil {
		t.Fatalf("NewWorkerContainer returned error: %v", err)
	}

	id, err := container.JobQueue.DequeueFormat(context.Background())
	if err != nil {
		t.Fatalf("dequeuing seed failed: %v", err)
	}
	if id != post.DarkPostID("post-local") {
		t.Fatalf("unexpected seed id: %s", id)
	}

	if err := container.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}
	if !stub.closed {
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
	t.Setenv("GEMINI_API_KEY", "dummy")
	t.Setenv("GEMINI_MODEL", "dummy")
	origRepoFactory := postRepositoryFactory
	postRepositoryFactory = func(ctx context.Context, infra *Infra) (repository.PostRepository, bool, error) {
		return nil, false, errors.New("repo factory error")
	}
	defer func() { postRepositoryFactory = origRepoFactory }()

	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when repository factory fails")
	}
}

func TestNewWorkerContainer_SeedPostsError(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "dummy")
	t.Setenv("GEMINI_MODEL", "dummy")
	origSeed := seedPostsFunc
	seedPostsFunc = func(ctx context.Context, repo repository.PostRepository) (post.DarkPostID, error) {
		return "", errors.New("seed error")
	}
	defer func() { seedPostsFunc = origSeed }()

	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when seeding posts fails")
	}
}

func TestNewWorkerContainer_FormatterFactoryError(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "dummy")
	t.Setenv("GEMINI_MODEL", "dummy")
	origFactory := formatterFactory
	formatterFactory = func(ctx context.Context) (llm.Formatter, func() error, error) {
		return nil, nil, errors.New("formatter error")
	}
	defer func() { formatterFactory = origFactory }()

	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when formatter factory fails")
	}
}

func TestNewWorkerContainer_MissingGeminiConfig(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GEMINI_MODEL", "")
	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when config is missing")
	}
}

func TestNewWorkerContainer_InfraError(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "dummy")
	t.Setenv("GEMINI_MODEL", "dummy")
	origInfra := infraFactory
	infraFactory = func(ctx context.Context) (*Infra, error) {
		return nil, errors.New("infra error")
	}
	defer func() { infraFactory = origInfra }()

	if _, err := NewWorkerContainer(context.Background()); err == nil {
		t.Fatalf("expected error when infra initialization fails")
	}
}

func TestNewWorkerContainer_SeedEnqueueError(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "dummy")
	t.Setenv("GEMINI_MODEL", "dummy")
	stub := &stubFormatter{}
	origFactory := formatterFactory
	formatterFactory = func(ctx context.Context) (llm.Formatter, func() error, error) {
		return stub, stub.Close, nil
	}
	defer func() { formatterFactory = origFactory }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	container, err := NewWorkerContainer(ctx)
	if err != nil {
		t.Fatalf("expected success even when enqueue fails: %v", err)
	}
	_ = container.Close()
}

func TestNewPostRepository_FirestoreRequiresClient(t *testing.T) {
	t.Setenv("WORKER_POST_REPOSITORY", "firestore")
	if _, _, err := newPostRepository(context.Background(), &Infra{}); err == nil {
		t.Fatalf("expected error when firestore client is missing")
	}
}

func TestNewPostRepository_FirestoreSuccess(t *testing.T) {
	t.Setenv("WORKER_POST_REPOSITORY", "firestore")
	infra := &Infra{firestoreClient: &firestore.Client{}}
	if _, seed, err := newPostRepository(context.Background(), infra); err != nil || seed {
		t.Fatalf("expected firestore repo without seeding, err=%v seed=%v", err, seed)
	}
}

func TestNewPostRepository_DefaultMemory(t *testing.T) {
	t.Setenv("WORKER_POST_REPOSITORY", "")
	repo, seed, err := newPostRepository(context.Background(), &Infra{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !seed {
		t.Fatalf("expected seed flag to be true for memory repo")
	}
	if _, ok := repo.(*repoMemory.InMemoryPostRepository); !ok {
		t.Fatalf("expected memory repository, got %T", repo)
	}
}

func TestSeedPosts_Error(t *testing.T) {
	repo := failingPostRepository{createErr: errors.New("create failed")}
	if _, err := seedPosts(context.Background(), repo); err == nil {
		t.Fatalf("expected error when create fails")
	}
}

func TestSeedPosts_SampleFactoryError(t *testing.T) {
	orig := samplePostFactory
	samplePostFactory = func() (*post.Post, error) {
		return nil, errors.New("sample error")
	}
	defer func() { samplePostFactory = orig }()

	if _, err := seedPosts(context.Background(), repoMemory.NewInMemoryPostRepository()); err == nil {
		t.Fatalf("expected error when sample factory fails")
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
var _ repository.PostRepository = (*failingPostRepository)(nil)

type failingPostRepository struct {
	createErr error
}

func (f failingPostRepository) Create(ctx context.Context, p *post.Post) error {
	return f.createErr
}

func (f failingPostRepository) Get(ctx context.Context, id post.DarkPostID) (*post.Post, error) {
	return nil, repository.ErrPostNotFound
}

func (f failingPostRepository) ListReady(ctx context.Context, limit int) ([]*post.Post, error) {
	return nil, nil
}

func (f failingPostRepository) Update(ctx context.Context, p *post.Post) error {
	return repository.ErrPostNotFound
}

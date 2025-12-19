package memory

import (
	"context"
	"sync"

	"backend/internal/domain/post"
	"backend/internal/port/repository"
)

// InMemoryPostRepository は PostRepository の簡易実装。
type InMemoryPostRepository struct {
	mu    sync.RWMutex
	store map[post.DarkPostID]*post.Post
}

// NewInMemoryPostRepository は InMemoryPostRepository を生成する。
func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{
		store: make(map[post.DarkPostID]*post.Post),
	}
}

func (r *InMemoryPostRepository) Create(ctx context.Context, p *post.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[p.ID()]; ok {
		return repository.ErrPostAlreadyExists
	}
	r.store[p.ID()] = p
	return nil
}

func (r *InMemoryPostRepository) Get(ctx context.Context, id post.DarkPostID) (*post.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.store[id]
	if !ok {
		return nil, repository.ErrPostNotFound
	}
	return p, nil
}

func (r *InMemoryPostRepository) ListReady(ctx context.Context, limit int) ([]*post.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*post.Post, 0, len(r.store))
	for _, p := range r.store {
		if p != nil && p.IsReady() {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *InMemoryPostRepository) Update(ctx context.Context, p *post.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[p.ID()]; !ok {
		return repository.ErrPostNotFound
	}
	r.store[p.ID()] = p
	return nil
}

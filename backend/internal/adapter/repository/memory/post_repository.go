package memory

import (
	"context"
	"sync"

	"backend/internal/domain/post"
	"backend/internal/port/repository"
)

// 簡易なメモリ常駐版の投稿リポジトリ。
type InMemoryPostRepository struct {
	mu    sync.RWMutex
	store map[post.DarkPostID]*post.Post
}

/**
 * 初期化済みマップを持つメモリリポジトリを返す。
 */
func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{
		store: make(map[post.DarkPostID]*post.Post),
	}
}

/**
 * 同じ ID が未登録であれば投稿を格納し、重複時はエラーにする。
 */
func (r *InMemoryPostRepository) Create(ctx context.Context, p *post.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// すでに登録済みなら重複エラーで返す
	if _, ok := r.store[p.ID()]; ok {
		return repository.ErrPostAlreadyExists
	}
	r.store[p.ID()] = p
	return nil
}

/**
 * ID で検索し、存在しなければ NotFound を返す。
 */
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
	count := 0
	for _, p := range r.store {
		// 公開待ちのみ返す
		if p != nil && p.IsReady() {
			result = append(result, p)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return result, nil
}

/**
 * 既存エントリのみ更新し、未登録なら NotFound を返す。
 */
func (r *InMemoryPostRepository) Update(ctx context.Context, p *post.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[p.ID()]; !ok {
		return repository.ErrPostNotFound
	}
	r.store[p.ID()] = p
	return nil
}

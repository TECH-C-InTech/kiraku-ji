package memory

import (
	"context"
	"errors"
	"sync"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/repository"
)

var (
	errNilDraw     = errors.New("memoryrepository: draw is nil")
	errEmptyPostID = errors.New("memoryrepository: post id is empty")
)

// InMemoryDrawRepository はメモリ上で Draw を管理するリポジトリ。
type InMemoryDrawRepository struct {
	mu    sync.RWMutex
	store map[post.DarkPostID]*drawdomain.Draw
}

// NewInMemoryDrawRepository は InMemoryDrawRepository を生成する。
func NewInMemoryDrawRepository() *InMemoryDrawRepository {
	return &InMemoryDrawRepository{
		store: make(map[post.DarkPostID]*drawdomain.Draw),
	}
}

// Create は Draw を保存する。既存 ID があれば ErrDrawAlreadyExists を返す。
func (r *InMemoryDrawRepository) Create(ctx context.Context, d *drawdomain.Draw) error {
	if d == nil {
		return errNilDraw
	}
	postID := d.PostID()
	if postID == "" {
		return errEmptyPostID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.store[postID]; exists {
		return repository.ErrDrawAlreadyExists
	}
	r.store[postID] = cloneDraw(d)
	return nil
}

// GetByPostID は指定した Post ID の Draw を返す。
func (r *InMemoryDrawRepository) GetByPostID(ctx context.Context, postID post.DarkPostID) (*drawdomain.Draw, error) {
	if postID == "" {
		return nil, repository.ErrDrawNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	d, ok := r.store[postID]
	if !ok || d == nil {
		return nil, repository.ErrDrawNotFound
	}

	return cloneDraw(d), nil
}

// ListReady は Verified な Draw をすべて返す。
func (r *InMemoryDrawRepository) ListReady(ctx context.Context) ([]*drawdomain.Draw, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*drawdomain.Draw, 0, len(r.store))
	for _, d := range r.store {
		if d == nil {
			continue
		}
		if d.Status() == drawdomain.StatusVerified {
			result = append(result, cloneDraw(d))
			continue
		}
		if d.Status() == drawdomain.StatusPending || d.Status() == drawdomain.StatusRejected {
			continue
		}
	}

	return result, nil
}

func cloneDraw(d *drawdomain.Draw) *drawdomain.Draw {
	if d == nil {
		return nil
	}
	clone := *d
	return &clone
}

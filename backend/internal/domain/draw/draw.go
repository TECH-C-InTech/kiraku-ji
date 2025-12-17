package draw

import (
	"errors"

	"backend/internal/domain/post"
)

var (
	// ErrEmptyResult は結果が空の場合に返される。
	ErrEmptyResult = errors.New("draw: result is empty")
	// ErrEmptyPostID は Post ID が空の場合に返される。
	ErrEmptyPostID = errors.New("draw: post id is empty")
	// ErrNilPost は nil の Post を受け取った際に返される。
	ErrNilPost = errors.New("draw: nil post supplied")
	// ErrPostNotReady は ready でない Post から Draw を生成しようとした際に返される。
	ErrPostNotReady = errors.New("draw: post is not ready")
)

// Draw はおみくじ結果を表す。
type Draw struct {
	postID string
	result string
}

// New creates a Draw for the given post ID and result.
// It returns ErrEmptyPostID if postID is empty and ErrEmptyResult if result is empty.
func New(postID, result string) (*Draw, error) {
	if postID == "" {
		return nil, ErrEmptyPostID
	}
	if result == "" {
		return nil, ErrEmptyResult
	}

	return &Draw{
		postID: postID,
		result: result,
	}, nil
}

// FromPost creates a Draw for the given Post using the Post's ID and the supplied result.
// It returns ErrNilPost if p is nil, ErrPostNotReady if p.IsReady() is false, or a validation error if the provided result is empty.
func FromPost(p *post.Post, result string) (*Draw, error) {
	if p == nil {
		return nil, ErrNilPost
	}
	if !p.IsReady() {
		return nil, ErrPostNotReady
	}

	return New(p.ID(), result)
}

// PostID は元となった Post の ID を返す。
func (d *Draw) PostID() string {
	return d.postID
}

// Result はおみくじ結果の本文を返す。
func (d *Draw) Result() string {
	return d.result
}
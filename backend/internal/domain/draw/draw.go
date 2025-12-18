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

type (
	FormattedContent string
	Status           string
)

// Status の種類
const (
	StatusPending  Status = "pending"
	StatusVerified Status = "verified"
	StatusRejected Status = "rejected"
)

// Draw はおみくじ結果を表す。
type Draw struct {
	postID post.DarkPostID
	result FormattedContent
	status Status
}

// New は Post ID と結果から Draw を生成する。
func New(postID post.DarkPostID, result FormattedContent) (*Draw, error) {
	if postID == "" {
		return nil, ErrEmptyPostID
	}
	if result == "" {
		return nil, ErrEmptyResult
	}

	return &Draw{
		postID: postID,
		result: result,
		status: StatusPending,
	}, nil
}

// FromPost は ready な Post から Draw を生成する。
func FromPost(p *post.Post, result FormattedContent) (*Draw, error) {
	if p == nil {
		return nil, ErrNilPost
	}
	if !p.IsReady() {
		return nil, ErrPostNotReady
	}

	return New(p.ID(), result)
}

// PostID は元となった Post の ID を返す。
func (d *Draw) PostID() post.DarkPostID {
	return d.postID
}

// Result はおみくじ結果の本文を返す。
func (d *Draw) Result() FormattedContent {
	return d.result
}

// Status はおみくじ結果の状態を返す。
func (d *Draw) Status() Status {
	return d.status
}

// MarkVerified は結果を検証済み状態へ遷移させる。
func (d *Draw) MarkVerified() {
	d.status = StatusVerified
}

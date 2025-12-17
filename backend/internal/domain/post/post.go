package post

import "errors"

// Status は闇投稿の状態を表す。
type Status string

const (
	StatusPending Status = "pending"
	StatusReady   Status = "ready"
)

var (
	// ErrEmptyContent は投稿内容が空の場合に返される。
	ErrEmptyContent = errors.New("post: content is empty")
	// ErrInvalidStatus は不正な状態が指定された際に返される。
	ErrInvalidStatus = errors.New("post: invalid status")
	// ErrInvalidStatusTransition は許可されていない状態遷移が要求された際に返される。
	ErrInvalidStatusTransition = errors.New("post: invalid status transition")
)

// Post は闇投稿そのもの。
type Post struct {
	id      string
	content string
	status  Status
}

// New creates a Post with the given id and content and sets its status to StatusPending.
// It returns ErrEmptyContent if content is empty.
func New(id, content string) (*Post, error) {
	if content == "" {
		return nil, ErrEmptyContent
	}

	return &Post{
		id:      id,
		content: content,
		status:  StatusPending,
	}, nil
}

// Restore reconstructs an existing Post using the given id, content, and status.
// It returns ErrEmptyContent when content is empty and ErrInvalidStatus when the
// provided status is not a recognized Status.
func Restore(id, content string, status Status) (*Post, error) {
	if content == "" {
		return nil, ErrEmptyContent
	}
	if !status.isValid() {
		return nil, ErrInvalidStatus
	}

	return &Post{
		id:      id,
		content: content,
		status:  status,
	}, nil
}

// ID は投稿の識別子を返す。
func (p *Post) ID() string {
	return p.id
}

// Content は投稿内容を返す。
func (p *Post) Content() string {
	return p.content
}

// Status は現在の状態を返す。
func (p *Post) Status() Status {
	return p.status
}

// IsReady は ready 状態かどうかを返す。
func (p *Post) IsReady() bool {
	return p.status == StatusReady
}

// MarkReady は pending -> ready の状態遷移のみを許可する。
func (p *Post) MarkReady() error {
	if p.status != StatusPending {
		return ErrInvalidStatusTransition
	}

	p.status = StatusReady
	return nil
}

func (s Status) isValid() bool {
	return s == StatusPending || s == StatusReady
}
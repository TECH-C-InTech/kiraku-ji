package post

import "errors"

type (
	// 闇投稿を一意に識別する ID。
	DarkPostID string
	// 整形前本文
	DarkContent string
	// 闇投稿の状態
	Status string
)

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
	id      DarkPostID
	content DarkContent
	status  Status
}

// New は新しい闇投稿を pending 状態で作成する。
func New(id DarkPostID, content DarkContent) (*Post, error) {
	if content == "" {
		return nil, ErrEmptyContent
	}

	return &Post{
		id:      id,
		content: content,
		status:  StatusPending,
	}, nil
}

// Restore は既存の投稿を再構築する。
func Restore(id DarkPostID, content DarkContent, status Status) (*Post, error) {
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
func (p *Post) ID() DarkPostID {
	return p.id
}

// Content は投稿内容を返す。
func (p *Post) Content() DarkContent {
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

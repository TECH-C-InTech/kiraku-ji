package llm

import (
	"context"
	"errors"

	"backend/internal/domain/draw"
	"backend/internal/domain/post"
)

var (
	ErrFormatterUnavailable = errors.New("llm: 整形サービスに接続できません")
	ErrContentRejected      = errors.New("llm: 投稿内容が拒否されました")
	ErrInvalidFormat        = errors.New("llm: 期待する形式で出力されませんでした")
)

/**
 * LLM にリクエストする際のデータ
 * @param DarkPostID 闇投稿 ID
 * @param DarkContent 整形対象の本文
 */
type FormatRequest struct {
	DarkPostID  post.DarkPostID
	DarkContent post.DarkContent
}

/**
 * LLM から返される整形済みデータ
 * @param DarkPostID 闇投稿 ID
 * @param FormattedContent 整形後の本文
 * @param Status 整形結果の状態
 * @param ValidationReason 検証理由（Status が Rejected の場合にセットされる）
 */
type FormatResult struct {
	DarkPostID       post.DarkPostID
	FormattedContent draw.FormattedContent
	Status           draw.Status
	ValidationReason string
}

/**
 * LLM フォーマッターの契約
 * Format: 常に StatusPending を返す整形処理
 * Validate: FormatResult を検証し、StatusVerified / StatusRejected をセットした結果を返す
 */
type Formatter interface {
	Format(ctx context.Context, req *FormatRequest) (*FormatResult, error)
	Validate(ctx context.Context, result *FormatResult) (*FormatResult, error)
}

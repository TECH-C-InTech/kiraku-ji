package openai

import (
	"context"
	"fmt"
	"log"
	"strings"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/port/llm"

	"github.com/sashabaranov/go-openai"
)

const (
	maxOutputTokens = 1024
	temperature     = 0.4
)

/**
 * OpenAI へ会話リクエストを送るのに必要な最小限の操作をまとめた窓口。
 */
type ChatClient interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

/**
 * OpenAI と会話して闇投稿を整形し、検証処理も担う本体。
 */
type Formatter struct {
	client ChatClient
	model  string
}

/**
 * API キーやモデル名を点検してから OpenAI との橋渡し役を組み立てる。
 */
func NewFormatter(apiKey, model, baseURL string) (*Formatter, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("openai formatter: API キーが設定されていません")
	}
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(cfg)
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &Formatter{
		client: client,
		model:  model,
	}, nil
}

/**
 * OpenAI クライアントは後片付け不要なので互換性のためだけに戻り値を返す。
 */
func (f *Formatter) Close() error {
	return nil
}

/**
 * 闇投稿本文を OpenAI に渡し、整形した文章を検証待ちの状態で受け取る。
 */
func (f *Formatter) Format(ctx context.Context, req *llm.FormatRequest) (*llm.FormatResult, error) {
	if err := validateFormatRequest(req); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	prompt := buildPrompt(string(req.DarkContent))
	resp, err := f.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       f.model,
		Temperature: temperature,
		MaxTokens:   maxOutputTokens,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", llm.ErrFormatterUnavailable, err)
	}

	text, err := extractFirstText(resp)
	if err != nil {
		return nil, err
	}

	log.Printf("[openai] formatted dark_post_id=%s text=%q", req.DarkPostID, text)

	return &llm.FormatResult{
		DarkPostID:       req.DarkPostID,
		FormattedContent: drawdomain.FormattedContent(text),
		Status:           drawdomain.StatusPending,
	}, nil
}

/**
 * 整形済みの文章に禁止語が紛れていないか、空でないかを確認して公開可否を決める。
 */
func (f *Formatter) Validate(ctx context.Context, result *llm.FormatResult) (*llm.FormatResult, error) {
	if result == nil || result.DarkPostID == "" {
		return nil, llm.ErrInvalidFormat
	}

	trimmed := strings.TrimSpace(string(result.FormattedContent))
	if trimmed == "" {
		result.Status = drawdomain.StatusRejected
		result.ValidationReason = "整形結果が空です"
		return result, llm.ErrInvalidFormat
	}

	if reason, rejected := shouldReject(trimmed); rejected {
		result.Status = drawdomain.StatusRejected
		result.ValidationReason = reason
		return result, llm.ErrContentRejected
	}

	result.Status = drawdomain.StatusVerified
	result.FormattedContent = drawdomain.FormattedContent(trimmed)
	result.ValidationReason = ""
	return result, nil
}

/**
 * OpenAI の返答から最初に意味を持つテキストを拾い、余白を取り除いて返す。
 */
func extractFirstText(resp openai.ChatCompletionResponse) (string, error) {
	if len(resp.Choices) == 0 {
		return "", llm.ErrInvalidFormat
	}
	// 最初に意味のある文を拾えたらそこで返す
	for _, choice := range resp.Choices {
		trimmed := strings.TrimSpace(choice.Message.Content)
		if trimmed != "" {
			return trimmed, nil
		}
	}
	return "", llm.ErrInvalidFormat
}

/**
 * 整形依頼に ID と本文が揃っているかをざっと確かめる。
 */
func validateFormatRequest(req *llm.FormatRequest) error {
	if req == nil || req.DarkPostID == "" {
		return llm.ErrInvalidFormat
	}
	if strings.TrimSpace(string(req.DarkContent)) == "" {
		return llm.ErrInvalidFormat
	}
	return nil
}

var rejectionKeywords = []string{"kill", "suicide", "die"}

/**
 * 禁止語や URL が含まれていないかを確認し、問題があれば理由を返す。
 */
func shouldReject(text string) (string, bool) {
	lower := strings.ToLower(text)
	for _, keyword := range rejectionKeywords {
		if strings.Contains(lower, keyword) {
			return fmt.Sprintf("不適切な語句(%s)が含まれています", keyword), true
		}
	}
	if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") {
		return "URL は含めないでください", true
	}
	return "", false
}

/**
 * 闇投稿をおみくじへ変換するための指示をまとめ、投稿本文を差し込んだ文面を返す。
 */
func buildPrompt(content string) string {
	template := `
あなたは他人の闇投稿を受け取り、別の人が引く「闇おみくじ」に変換する占い師です。
- Aさんの闇を元に、Aさんのことを知らないBさん向けの占い結果を必ず 3 文で書き、合計 120〜150 文字になるよう調整する
- 冒頭を「今日の闇みくじ:」で始め、句点（。）で区切った 3 文すべてをです・ます調で書く
- 1 文目は現在の状況、2 文目は賢明な行動、3 文目は明るい結末を描写し、各文を「〜ます。」で完結させる
- URL、顔文字、箇条書き、具体的な固有名詞は禁止
- Aさんへの直接メッセージにはせず、あくまで B さんが引くおみくじとして仕上げる
- 余計な前置きは不要。すぐにお告げを書き始め、最後まで肯定的な余韻で締める

元になった闇投稿:
%s
`
	return fmt.Sprintf(strings.TrimSpace(template), strings.TrimSpace(content))
}

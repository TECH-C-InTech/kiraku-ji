package openai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"backend/internal/config"
	drawdomain "backend/internal/domain/draw"
	"backend/internal/port/llm"

	"github.com/sashabaranov/go-openai"
)

const (
	maxOutputTokens       = 1024
	temperature           = 0.4
	maxFormattedLength    = 150
	minFormattedLength    = 120
	fortunePrefix         = "今日のきらくじ:"
	expectedSentenceCount = 3
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
		model = config.DefaultOpenAIModel
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

	normalized := normalizeFortuneText(trimmed)
	if reason, rejected := shouldReject(normalized); rejected {
		result.Status = drawdomain.StatusRejected
		result.ValidationReason = reason
		return result, llm.ErrContentRejected
	}

	result.Status = drawdomain.StatusVerified
	result.FormattedContent = drawdomain.FormattedContent(normalized)
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
 * 文字数や構成、禁止語を確認し、問題があれば理由を返す。
 */
func shouldReject(text string) (string, bool) {
	length := utf8.RuneCountInString(text)
	if length < minFormattedLength {
		return "整形結果が短すぎます", true
	}
	if length > maxFormattedLength {
		return "整形結果が長すぎます", true
	}

	lower := strings.ToLower(text)
	for _, keyword := range rejectionKeywords {
		if strings.Contains(lower, keyword) {
			return fmt.Sprintf("不適切な語句(%s)が含まれています", keyword), true
		}
	}
	if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") {
		return "URL は含めないでください", true
	}
	if !strings.HasPrefix(text, fortunePrefix) {
		return fmt.Sprintf("冒頭は「%s」で始めてください", fortunePrefix), true
	}
	if reason, rejected := violatesFortuneStructure(text); rejected {
		return reason, true
	}
	return "", false
}

/**
 * 闇投稿をおみくじへ変換するための指示をまとめ、投稿本文を差し込んだ文面を返す。
 */
func buildPrompt(content string) string {
	template := `
あなたは他人の闇投稿をもとに、別の人が引く「きらくじ」を作るメンヘラ占い師です。出力は日本語のみで行い、次の指示を厳守してください。

【文章ルール】
1. 合計 120〜150 文字の 3 文構成で書く。
2. 各文の内容: (1) 今の状況は少し重めに捉える (2) 賢明な行動は具体的で粘り強く、ねちねちした現実的な対処 (3) 結末は少しユーモアを含めつつ、癒しになるような余韻を残す。
3. 3 文すべて「〜ます。」で終え、句点（。）で区切る。
4. 固有名詞・URL・箇条書き・顔文字は禁止
5. A さんへの直接メッセージにはせず、B さんが引くきらくじとして書く

【出力フォーマット】
今日のきらくじ: 一文目。二文目。三文目。
- 冒頭は必ず「今日のきらくじ:」ではじめ、余計な前置きや後書きは不要
- 改行せず 1 行で書ききる

上記ルールを完全に満たす文章だけを 1 行で返してください。

元になった闇投稿:
%s`
	return fmt.Sprintf(strings.TrimSpace(template), strings.TrimSpace(content))
}

/**
 * 改行や余白を整え、検証しやすい形へ揃える。
 */
func normalizeFortuneText(text string) string {
	noCR := strings.ReplaceAll(text, "\r", "")
	noLF := strings.ReplaceAll(noCR, "\n", "")
	return strings.TrimSpace(noLF)
}

/**
 * お告げ文の構成や語尾が条件を満たしているかを調べる。
 */
func violatesFortuneStructure(text string) (string, bool) {
	body := strings.TrimSpace(strings.TrimPrefix(text, fortunePrefix))
	sentences := splitSentences(body)
	if len(sentences) != expectedSentenceCount {
		return "お告げは3文構成で書いてください", true
	}
	for idx, sentence := range sentences {
		if !strings.HasSuffix(sentence, "ます") {
			return fmt.Sprintf("%d文目は「〜ます」で終えてください", idx+1), true
		}
	}
	return "", false
}

/**
 * 句点で区切った文を抽出し、空の要素を除いて返す。
 */
func splitSentences(body string) []string {
	raw := strings.Split(body, "。")
	sentences := make([]string, 0, len(raw))
	for _, part := range raw {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		sentences = append(sentences, trimmed)
	}
	return sentences
}

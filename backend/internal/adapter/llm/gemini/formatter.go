package gemini

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/port/llm"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	defaultModelName   = "gemini-2.5-flash"
	maxFormattedLength = 400
	minFormattedLength = 12
)

var rejectionKeywords = []string{"kill", "suicide", "die"}

var newGeminiClient = genai.NewClient

// Gemini の生成モデルをテスト用に差し替えやすくしたインターフェース。
type contentGenerator interface {
	GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// Gemini を用いた整形処理と検証処理をまとめたもの。
type Formatter struct {
	generator contentGenerator
	closeFn   func() error
	modelName string
}

/**
 * API キーなどの設定から Gemini への窓口を構築し、整形器を返す。
 * 必須情報が欠けていたり、接続ができないときはその旨を伝えて終了する。
 */
func NewFormatter(ctx context.Context, apiKey, modelName string, extraOpts ...option.ClientOption) (*Formatter, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("gemini formatter: API キーが設定されていません")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	opts := append([]option.ClientOption{option.WithAPIKey(apiKey)}, extraOpts...)
	client, err := newGeminiClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", llm.ErrFormatterUnavailable, err)
	}

	resolvedModel := resolveModelName(modelName)
	model := client.GenerativeModel(resolvedModel)
	configured := configureModel(model)

	return &Formatter{
		generator: configured,
		closeFn:   makeCloseFn(client),
		modelName: resolvedModel,
	}, nil
}

/**
 * 内部で保持している接続を後片付けする。
 * そもそも接続していない場合は何もせずに戻る。
 */
func (f *Formatter) Close() error {
	if f == nil || f.closeFn == nil {
		return nil
	}
	return f.closeFn()
}

/**
 * 闇投稿本文を丁寧な言葉へ整え、検証待ち状態の結果として返す。
 * 依頼が空だったり応答が壊れている場合は、理由を添えて失敗を知らせる。
 */
func (f *Formatter) Format(ctx context.Context, req *llm.FormatRequest) (*llm.FormatResult, error) {
	if err := validateFormatRequest(req); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if f.generator == nil {
		return nil, fmt.Errorf("%w: gemini formatter: 生成器が初期化されていません", llm.ErrFormatterUnavailable)
	}

	prompt := buildPrompt(string(req.DarkContent))
	resp, err := f.generator.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", llm.ErrFormatterUnavailable, err)
	}

	text, err := extractFirstText(resp)
	if err != nil {
		return nil, err
	}

	return &llm.FormatResult{
		DarkPostID:       req.DarkPostID,
		FormattedContent: drawdomain.FormattedContent(text),
		Status:           drawdomain.StatusPending,
	}, nil
}

/**
 * 整形結果が投稿規約に沿っているかを再確認し、公開可否を決める。
 * 禁止語や文字数違反などが見つかったら拒否理由を付けて返す。
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
 * 整形依頼に ID と本文が入っているかを確かめる。
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

/**
 * モデル名の指定が空だった場合、既定の名前へ置き換える。
 */
func resolveModelName(name string) string {
	if strings.TrimSpace(name) == "" {
		return defaultModelName
	}
	return name
}

/**
 * 整形時の言い回しや禁止事項を明記したガイド文を作り、投稿本文を差し込む。
 */
func buildPrompt(content string) string {
	template := `
あなたは匿名の悩み相談を受け取り、投稿者を肯定しながら穏やかで前向きな 200 字以内の日本語メッセージに整形する編集者です。
- です・ます調で丁寧に書く
- URL や顔文字、箇条書きは禁止
- 余計な前置きは書かず、すぐ本文を書き始める

原文:
%s
`
	return fmt.Sprintf(strings.TrimSpace(template), strings.TrimSpace(content))
}

/**
 * Gemini の応答候補から先頭の文章を取り出す。
 * 何も得られない場合は整形不備として扱う。
 */
func extractFirstText(resp *genai.GenerateContentResponse) (string, error) {
	if resp == nil {
		return "", llm.ErrInvalidFormat
	}
	for _, candidate := range resp.Candidates {
		if candidate == nil || candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if part == nil {
				continue
			}
			if text, ok := part.(genai.Text); ok {
				trimmed := strings.TrimSpace(string(text))
				if trimmed != "" {
					return trimmed, nil
				}
			}
		}
	}
	return "", llm.ErrInvalidFormat
}

/**
 * 文字数・禁止語・URL などの検査を行い、違反が見つかったら拒否理由を返す。
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
	return "", false
}

/**
 * 候補数・文字数上限・温度などの設定を行い、生成器として扱えるようにする。
 */
func configureModel(model *genai.GenerativeModel) contentGenerator {
	if model == nil {
		return nil
	}
	model.SetCandidateCount(1)
	model.SetMaxOutputTokens(512)
	model.SetTemperature(0.4)
	return model
}

/**
 * クライアントが存在する場合だけ後片付け用の関数を返す。
 * 無い場合は nil を返し、余計な Close を避ける。
 */
func makeCloseFn(client *genai.Client) func() error {
	if client == nil {
		return nil
	}
	return client.Close
}

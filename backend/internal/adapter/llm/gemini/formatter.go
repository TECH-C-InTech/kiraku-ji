package gemini

import (
	"context"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/port/llm"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	defaultModelName      = "gemini-2.5-flash"
	maxFormattedLength    = 150
	minFormattedLength    = 120
	fortunePrefix         = "今日のきらくじ:"
	expectedSentenceCount = 3
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

	log.Printf("[gemini] formatted dark_post_id=%s text=%q", req.DarkPostID, text)

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
あなたは他人の闇投稿をもとに、別の人が引く「きらくじ」を作る占い師です。出力は日本語のみで行い、次の指示を厳守してください。

【文章ルール】
1. 合計 120〜150 文字の 3 文構成で書く。
2. 各文の内容: (1) 今の状況は少し重めに捉える (2) 賢明な行動は具体的で粘り強く、ねちねちした現実的な対処 (3) 結末は少しユーモアを含めつつ、癒しになるような余韻を残す。
3. 3 文すべて「〜ます。」で終え、句点（。）で区切る。
4. 固有名詞・URL・箇条書き・顔文字は禁止
5. A さんへの直接メッセージにはせず、B さんが引くきらくじとして書く。

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
 * Gemini の応答候補から先頭の文章を取り出す。
 * 何も得られない場合は整形不備として扱う。
 */
func extractFirstText(resp *genai.GenerateContentResponse) (string, error) {
	if resp == nil {
		return "", llm.ErrInvalidFormat
	}
	for idx, candidate := range resp.Candidates {
		if candidate == nil || candidate.Content == nil {
			continue
		}
		var builder strings.Builder
		for pIdx, part := range candidate.Content.Parts {
			if part == nil {
				continue
			}
			if text, ok := part.(genai.Text); ok {
				builder.WriteString(string(text))
				log.Printf("[gemini debug] candidate=%d part=%d text=%q", idx, pIdx, string(text))
			}
		}
		trimmed := strings.TrimSpace(builder.String())
		if trimmed != "" {
			return trimmed, nil
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
	if !strings.HasPrefix(text, fortunePrefix) {
		return fmt.Sprintf("冒頭は「%s」で始めてください", fortunePrefix), true
	}
	if reason, rejected := violatesFortuneStructure(text); rejected {
		return reason, true
	}
	return "", false
}

func normalizeFortuneText(text string) string {
	noCR := strings.ReplaceAll(text, "\r", "")
	noLF := strings.ReplaceAll(noCR, "\n", "")
	return strings.TrimSpace(noLF)
}

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

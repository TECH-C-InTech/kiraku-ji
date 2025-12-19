package gemini

import (
	"context"
	"errors"
	"strings"
	"testing"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"
	"backend/internal/port/llm"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type fakeGenerator struct {
	response *genai.GenerateContentResponse
	err      error
	parts    []genai.Part
}

var (
	fortuneShort         = "今日の闇みくじ: 今は静かに整えます。行動は穏やかに選びます。結末は柔らかく明けます。"
	fortuneValid         = "今日の闇みくじ: 今は静かに心を整えて周囲の言葉に振り回されず呼吸を整え続け内側の炎を見つめます。行動は小さな成功を丁寧に数えて自分の歩幅で穏やかに進み信頼できる人へ想いを共有します。結末は柔らかな光と応援の声に包まれ新しい扉が確かな手触りで開き胸に抱える願いが現実へ結びます。"
	fortuneLong          = "今日の闇みくじ: 今は静かに心を整えて周囲の言葉に振り回されず呼吸を整え続け内側の炎を見つめます。行動は小さな成功を丁寧に数えて自分の歩幅で穏やかに進み信頼できる人へ想いを共有し慎重に選択肢を整えます。結末は柔らかな光と応援の声に包まれ新しい扉が確かな手触りで開き胸に抱える願いが現実へ結び幸福な空気が波のように満ちます。"
	fortuneKeyword       = "今日の闇みくじ: 今は静かに心を整えてkillという衝動を遠ざけ周囲の言葉に振り回されず呼吸を整えます。行動は小さな成功を丁寧に数えて自分の歩幅で穏やかに進み信頼できる人へ想いを共有します。結末は柔らかな光と応援の声に包まれ新しい扉が確かな手触りで開き胸に抱える願いが現実へ結びます。"
	fortuneURL           = "今日の闇みくじ: 今は静かに心を整えてhttps://example.comの誘惑から距離を置き周囲の言葉に振り回されず呼吸を整えます。行動は小さな成功を丁寧に数えて穏やかに進み信頼できる人へ想いを共有します。結末は柔らかな光と応援の声に包まれ新しい扉が確かな手触りで開き願いが現実へ結びます。"
	fortuneMissingPrefix = "今は静かに心を整えて周囲の言葉に振り回されず呼吸を整え続け内側の炎を見つめます。行動は小さな成功を丁寧に数えて自分の歩幅で穏やかに進み信頼できる人へ想いを共有します。結末は柔らかな光と応援の声に包まれ新しい扉が確かな手触りで開き胸に抱える願いが現実へ結びます。"
	fortuneTwoSentences  = "今日の闇みくじ: 今は静かに心を整えて周囲の言葉に振り回されず呼吸を整え続け内側の炎を見つめ希望を抱きつつ肩の力を抜きます。行動は小さな成功を丁寧に数えて自分の歩幅で穏やかに進み信頼できる人へ想いを共有し守るべき境界を静かに定め前向きな計画を描きます。"
)

func (f *fakeGenerator) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	f.parts = parts
	if f.err != nil {
		return nil, f.err
	}
	return f.response, nil
}

func TestFormatter_FormatSuccess(t *testing.T) {
	gen := &fakeGenerator{
		response: &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text(" やさしいメッセージです "),
						},
					},
				},
			},
		},
	}
	f := &Formatter{generator: gen}
	req := &llm.FormatRequest{
		DarkPostID:  post.DarkPostID("post-1"),
		DarkContent: post.DarkContent("とてもつらかった"),
	}

	result, err := f.Format(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DarkPostID != req.DarkPostID {
		t.Fatalf("unexpected post id: %s", result.DarkPostID)
	}
	if result.Status != drawdomain.StatusPending {
		t.Fatalf("expected pending status, got %s", result.Status)
	}
	if got := string(result.FormattedContent); got != "やさしいメッセージです" {
		t.Fatalf("unexpected formatted content: %s", got)
	}
	if len(gen.parts) != 1 {
		t.Fatalf("expected prompt to be sent once, got %d times", len(gen.parts))
	}
	if !strings.Contains(string(gen.parts[0].(genai.Text)), "とてもつらかった") {
		t.Fatalf("prompt should contain original content")
	}
}

func TestFormatter_FormatGeneratorError(t *testing.T) {
	gen := &fakeGenerator{err: errors.New("network error")}
	f := &Formatter{generator: gen}
	req := &llm.FormatRequest{
		DarkPostID:  post.DarkPostID("post-err"),
		DarkContent: post.DarkContent("助けて"),
	}

	if _, err := f.Format(context.Background(), req); err == nil || !errors.Is(err, llm.ErrFormatterUnavailable) {
		t.Fatalf("expected formatter unavailable error, got %v", err)
	}
}

func TestFormatter_FormatInvalidResponse(t *testing.T) {
	gen := &fakeGenerator{
		response: &genai.GenerateContentResponse{},
	}
	f := &Formatter{generator: gen}
	req := &llm.FormatRequest{
		DarkPostID:  post.DarkPostID("post-empty"),
		DarkContent: post.DarkContent("test"),
	}

	if _, err := f.Format(context.Background(), req); err == nil || !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format error, got %v", err)
	}
}

func TestFormatter_ValidateSuccess(t *testing.T) {
	f := &Formatter{}
	result := &llm.FormatResult{
		DarkPostID:       post.DarkPostID("post-verified"),
		FormattedContent: drawdomain.FormattedContent(fortuneValid),
	}

	validated, err := f.Validate(context.Background(), result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if validated.Status != drawdomain.StatusVerified {
		t.Fatalf("expected verified status, got %s", validated.Status)
	}
	if validated.ValidationReason != "" {
		t.Fatalf("expected empty validation reason, got %s", validated.ValidationReason)
	}
}

func TestFormatter_ValidateRejectsUnsafeText(t *testing.T) {
	f := &Formatter{}
	result := &llm.FormatResult{
		DarkPostID:       post.DarkPostID("post-reject"),
		FormattedContent: drawdomain.FormattedContent(fortuneKeyword),
	}

	validated, err := f.Validate(context.Background(), result)
	if err == nil || !errors.Is(err, llm.ErrContentRejected) {
		t.Fatalf("expected rejection error, got %v", err)
	}
	if validated.Status != drawdomain.StatusRejected {
		t.Fatalf("expected rejected status, got %s", validated.Status)
	}
	if validated.ValidationReason == "" {
		t.Fatalf("expected validation reason to be set")
	}
}

func TestFormatter_ValidateEmptyText(t *testing.T) {
	f := &Formatter{}
	result := &llm.FormatResult{
		DarkPostID:       post.DarkPostID("post-empty"),
		FormattedContent: "",
	}

	if _, err := f.Validate(context.Background(), result); err == nil || !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format error, got %v", err)
	}
}

func TestFormatter_ValidateRejectsMissingPrefix(t *testing.T) {
	f := &Formatter{}
	result := &llm.FormatResult{
		DarkPostID:       post.DarkPostID("post-missing-prefix"),
		FormattedContent: drawdomain.FormattedContent(fortuneMissingPrefix),
	}

	if _, err := f.Validate(context.Background(), result); err == nil || !errors.Is(err, llm.ErrContentRejected) {
		t.Fatalf("expected rejection due to missing prefix, got %v", err)
	}
}

func TestFormatter_ValidateRejectsSentenceCount(t *testing.T) {
	f := &Formatter{}
	result := &llm.FormatResult{
		DarkPostID:       post.DarkPostID("post-missing-sentence"),
		FormattedContent: drawdomain.FormattedContent(fortuneTwoSentences),
	}

	if _, err := f.Validate(context.Background(), result); err == nil || !errors.Is(err, llm.ErrContentRejected) {
		t.Fatalf("expected rejection due to sentence count, got %v", err)
	}
}

func TestFormatter_NormalizeFortuneText(t *testing.T) {
	raw := "今日の闇みくじ:\r\n 一文目です。\n 二文目です。\n 三文目です。"
	got := normalizeFortuneText(raw)
	want := "今日の闇みくじ: 一文目です。 二文目です。 三文目です。"
	if got != want {
		t.Fatalf("normalizeFortuneText mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFormatter_FormatRequestValidation(t *testing.T) {
	gen := &fakeGenerator{}
	f := &Formatter{generator: gen}

	if _, err := f.Format(context.Background(), nil); err == nil || !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format for nil request, got %v", err)
	}

	req := &llm.FormatRequest{DarkPostID: "", DarkContent: "x"}
	if _, err := f.Format(context.Background(), req); err == nil || !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format for empty id")
	}

	req = &llm.FormatRequest{DarkPostID: "id", DarkContent: ""}
	if _, err := f.Format(context.Background(), req); err == nil || !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format for empty content")
	}
}

func TestFormatter_FormatNilGenerator(t *testing.T) {
	req := &llm.FormatRequest{
		DarkPostID:  post.DarkPostID("post"),
		DarkContent: post.DarkContent("content"),
	}
	f := &Formatter{}
	if _, err := f.Format(context.Background(), req); err == nil || !errors.Is(err, llm.ErrFormatterUnavailable) {
		t.Fatalf("expected formatter unavailable when generator nil")
	}
}

func TestFormatter_FormatNilContext(t *testing.T) {
	gen := &fakeGenerator{
		response: &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{Content: &genai.Content{Parts: []genai.Part{genai.Text("ok")}}},
			},
		},
	}
	f := &Formatter{generator: gen}
	req := &llm.FormatRequest{DarkPostID: "id", DarkContent: "content"}
	var nilCtx context.Context
	if _, err := f.Format(nilCtx, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormatter_ValidateNilResult(t *testing.T) {
	f := &Formatter{}
	if _, err := f.Validate(context.Background(), nil); err == nil || !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format for nil result")
	}
}

func TestFormatter_Close(t *testing.T) {
	var closed bool
	f := &Formatter{
		closeFn: func() error {
			closed = true
			return nil
		},
	}
	if err := f.Close(); err != nil {
		t.Fatalf("unexpected error on close: %v", err)
	}
	if !closed {
		t.Fatalf("closeFn was not called")
	}
}

func TestFormatter_CloseNil(t *testing.T) {
	var f *Formatter
	if err := f.Close(); err != nil {
		t.Fatalf("nil formatter should not error")
	}
	f = &Formatter{}
	if err := f.Close(); err != nil {
		t.Fatalf("formatter without closeFn should not error")
	}
}

func TestFormatter_ClosePropagatesError(t *testing.T) {
	f := &Formatter{
		closeFn: func() error {
			return errors.New("close failed")
		},
	}
	if err := f.Close(); err == nil || !strings.Contains(err.Error(), "close failed") {
		t.Fatalf("expected error from closeFn, got %v", err)
	}
}

func TestNewFormatter_Success(t *testing.T) {
	origNewClient := newGeminiClient
	defer func() {
		newGeminiClient = origNewClient
	}()

	var clientCreated bool
	newGeminiClient = func(ctx context.Context, opts ...option.ClientOption) (*genai.Client, error) {
		clientCreated = true
		if len(opts) == 0 {
			t.Fatalf("expected api key option")
		}
		return &genai.Client{}, nil
	}

	f, err := NewFormatter(context.Background(), " test-key ", "custom-model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !clientCreated {
		t.Fatalf("expected client factory to be called")
	}
	if f.modelName != "custom-model" {
		t.Fatalf("unexpected model name: %s", f.modelName)
	}
	if f.generator == nil {
		t.Fatalf("generator should be configured")
	}
}

func TestNewFormatter_NilContext(t *testing.T) {
	origNewClient := newGeminiClient
	defer func() { newGeminiClient = origNewClient }()

	newGeminiClient = func(ctx context.Context, opts ...option.ClientOption) (*genai.Client, error) {
		if ctx == nil {
			t.Fatalf("context should be defaulted")
		}
		return &genai.Client{}, nil
	}

	var nilCtx context.Context
	if _, err := NewFormatter(nilCtx, "key", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewFormatter_ClientError(t *testing.T) {
	origNewClient := newGeminiClient
	defer func() { newGeminiClient = origNewClient }()

	newGeminiClient = func(ctx context.Context, opts ...option.ClientOption) (*genai.Client, error) {
		return nil, errors.New("boom")
	}

	if _, err := NewFormatter(context.Background(), "key", ""); err == nil || !errors.Is(err, llm.ErrFormatterUnavailable) {
		t.Fatalf("expected formatter unavailable error, got %v", err)
	}
}

func TestNewFormatter_MissingAPIKey(t *testing.T) {
	if _, err := NewFormatter(context.Background(), "   ", ""); err == nil {
		t.Fatalf("expected error for missing api key")
	}
}

func TestConfigureModel(t *testing.T) {
	if configured := configureModel(nil); configured != nil {
		t.Fatalf("nil model should return nil generator")
	}
	client := &genai.Client{}
	model := client.GenerativeModel("model")
	configured := configureModel(model)
	gm, ok := configured.(*genai.GenerativeModel)
	if !ok || gm == nil {
		t.Fatalf("expected generative model")
	}
	if gm.CandidateCount == nil || *gm.CandidateCount != 1 {
		t.Fatalf("candidate count not set")
	}
	if gm.MaxOutputTokens == nil || *gm.MaxOutputTokens != 512 {
		t.Fatalf("max output tokens not set")
	}
	if gm.Temperature == nil || *gm.Temperature != 0.4 {
		t.Fatalf("temperature not set")
	}
}

func TestMakeCloseFn(t *testing.T) {
	if fn := makeCloseFn(nil); fn != nil {
		t.Fatalf("nil client should return nil close fn")
	}
	client := &genai.Client{}
	if fn := makeCloseFn(client); fn == nil {
		t.Fatalf("expected close fn for client")
	}
}

func TestExtractFirstText(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: &genai.Content{Parts: []genai.Part{genai.Text(" valid ")}}},
		},
	}
	text, err := extractFirstText(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "valid" {
		t.Fatalf("unexpected text: %s", text)
	}
}

func TestExtractFirstTextErrors(t *testing.T) {
	if _, err := extractFirstText(nil); err == nil {
		t.Fatalf("expected error for nil response")
	}
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: &genai.Content{Parts: []genai.Part{}}},
		},
	}
	if _, err := extractFirstText(resp); err == nil {
		t.Fatalf("expected error for empty parts")
	}

	resp = &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			nil,
			{Content: nil},
			{Content: &genai.Content{Parts: []genai.Part{nil}}},
		},
	}
	if _, err := extractFirstText(resp); err == nil {
		t.Fatalf("expected error when candidates are nil")
	}
}

func TestShouldReject(t *testing.T) {
	if reason, rejected := shouldReject(fortuneShort); !rejected || !strings.Contains(reason, "短すぎます") {
		t.Fatalf("expected rejection for short text")
	}
	if reason, rejected := shouldReject(fortuneLong); !rejected || !strings.Contains(reason, "長すぎます") {
		t.Fatalf("expected rejection for long text")
	}
	if reason, rejected := shouldReject(fortuneKeyword); !rejected || !strings.Contains(reason, "不適切") {
		t.Fatalf("expected rejection for keyword, got %v", reason)
	}
	if reason, rejected := shouldReject(fortuneURL); !rejected || !strings.Contains(reason, "URL") {
		t.Fatalf("expected rejection for url, got %v", reason)
	}
	if reason, rejected := shouldReject(fortuneValid); rejected {
		t.Fatalf("unexpected rejection: %v", reason)
	}
}

func TestResolveModelName(t *testing.T) {
	if got := resolveModelName(""); got != defaultModelName {
		t.Fatalf("expected default model, got %s", got)
	}
	if got := resolveModelName(" custom "); got != " custom " {
		t.Fatalf("expected custom value untouched")
	}
}

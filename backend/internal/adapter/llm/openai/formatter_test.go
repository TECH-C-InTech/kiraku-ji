package openai

import (
	"context"
	"errors"
	"strings"
	"testing"

	"backend/internal/config"
	drawdomain "backend/internal/domain/draw"
	"backend/internal/port/llm"

	githubOpenAI "github.com/sashabaranov/go-openai"
)

type stubChatClient struct {
	resp        githubOpenAI.ChatCompletionResponse
	err         error
	capturedCtx context.Context
	capturedReq githubOpenAI.ChatCompletionRequest
}

var (
	fortuneShort         = "今日のきらくじ: 胸の奥が重く曇ります。反応を待ちながら様子を見ます。最後は少し笑えて癒されます。"
	fortuneValid         = "今日のきらくじ: 心の奥がじっと湿って、気になる言葉が何度も頭に残り、寝不足も続いています。ひとつずつ事実を確認し、記録を残して、反応を待ちながら淡々と片付け、手順を崩さず進めます。最後には執念が効いて小さな勝ちを拾えたと笑え、ふっと癒されます。"
	fortuneLong          = "今日のきらくじ: 心の奥がじっと湿って、気になる言葉が何度も頭に残り、寝不足も続き、ため息が増えています。ひとつずつ事実を確認し、記録を残して、反応を待ちながら淡々と片付け、手順を崩さず進め、証拠の順序も丁寧に整えます。最後には執念が効いて小さな勝ちを拾えたと笑え、ふっと癒される余韻がしばらく長く残ります。"
	fortuneKeyword       = "今日のきらくじ: 心の奥がじっと湿って、気になる言葉が何度も頭に残り、killという語がちらついています。ひとつずつ事実を確認し、記録を残して、反応を待ちながら淡々と片付けます。最後には執念が効いて小さな勝ちを拾えたと笑え、ふっと癒されます。"
	fortuneURL           = "今日のきらくじ: 心の奥がじっと湿って、気になる言葉が何度も頭に残り、https://example.comの通知が気になります。ひとつずつ事実を確認し、記録を残して、反応を待ちながら淡々と片付けます。最後には執念が効いて小さな勝ちを拾えたと笑え、ふっと癒されます。"
	fortuneMissingPrefix = "心の奥がじっと湿って、気になる言葉が何度も頭に残り、寝不足も少し続いて眠りも浅くなっています。ひとつずつ事実を確認し、記録を残して、反応を待ちながら淡々と片付け、手順を崩さず進めます。最後には執念が効いて小さな勝ちを拾えたと笑え、ふっと癒されます。"
	fortuneTwoSentences  = "今日のきらくじ: 心の奥がじっと湿って、気になる言葉が何度も頭に残り、寝不足も続いています。ひとつずつ事実を確認し、記録を残して、反応を待ちながら淡々と片付け、手順を崩さず黙々と進め、返信の時間も決め、痕跡を整え、最後は少し笑えて癒されます。"
)

func (s *stubChatClient) CreateChatCompletion(ctx context.Context, req githubOpenAI.ChatCompletionRequest) (githubOpenAI.ChatCompletionResponse, error) {
	s.capturedCtx = ctx
	s.capturedReq = req
	return s.resp, s.err
}

func TestFormatterFormatSuccess(t *testing.T) {
	client := &stubChatClient{
		resp: githubOpenAI.ChatCompletionResponse{
			Choices: []githubOpenAI.ChatCompletionChoice{{
				Message: githubOpenAI.ChatCompletionMessage{Content: " 整形済み "},
			}},
		},
	}
	f := &Formatter{client: client, model: "test-model"}

	res, err := f.Format(context.Background(), &llm.FormatRequest{
		DarkPostID:  "post-1",
		DarkContent: "闇",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(res.FormattedContent) != "整形済み" {
		t.Fatalf("unexpected content: %q", res.FormattedContent)
	}
	if res.Status != "pending" {
		t.Fatalf("expected pending status, got %s", res.Status)
	}
}

func TestFormatterFormatClientError(t *testing.T) {
	client := &stubChatClient{err: errors.New("boom")}
	f := &Formatter{client: client, model: "test"}

	_, err := f.Format(context.Background(), &llm.FormatRequest{DarkPostID: "p", DarkContent: "x"})
	if !errors.Is(err, llm.ErrFormatterUnavailable) {
		t.Fatalf("expected formatter unavailable, got %v", err)
	}
}

func TestFormatterFormatInvalidResponse(t *testing.T) {
	client := &stubChatClient{resp: githubOpenAI.ChatCompletionResponse{Choices: nil}}
	f := &Formatter{client: client, model: "test"}

	_, err := f.Format(context.Background(), &llm.FormatRequest{DarkPostID: "p", DarkContent: "x"})
	if !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format, got %v", err)
	}
}

func TestFormatterFormatInvalidRequest(t *testing.T) {
	f := &Formatter{client: &stubChatClient{}, model: "test"}
	_, err := f.Format(context.Background(), &llm.FormatRequest{})
	if !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format, got %v", err)
	}
}

func TestFormatterFormatNilRequest(t *testing.T) {
	f := &Formatter{client: &stubChatClient{}, model: "test"}
	if _, err := f.Format(context.Background(), nil); !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format for nil request, got %v", err)
	}
}

func TestFormatterFormatDefaultsContext(t *testing.T) {
	client := &stubChatClient{
		resp: githubOpenAI.ChatCompletionResponse{
			Choices: []githubOpenAI.ChatCompletionChoice{{
				Message: githubOpenAI.ChatCompletionMessage{Content: "ok"},
			}},
		},
	}
	f := &Formatter{client: client, model: "test"}

	if _, err := f.Format(context.TODO(), &llm.FormatRequest{DarkPostID: "id", DarkContent: "text"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.capturedCtx == nil {
		t.Fatalf("expected context to be defaulted")
	}
	if client.capturedReq.Model != "test" {
		t.Fatalf("model not set in request: %v", client.capturedReq.Model)
	}
}

func TestFormatterValidate(t *testing.T) {
	f := &Formatter{}
	result, err := f.Validate(context.Background(), &llm.FormatResult{
		DarkPostID:       "post",
		FormattedContent: drawdomain.FormattedContent(fortuneValid),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "verified" {
		t.Fatalf("expected verified, got %s", result.Status)
	}
}

func TestFormatterValidateRejects(t *testing.T) {
	f := &Formatter{}
	result, err := f.Validate(context.Background(), &llm.FormatResult{
		DarkPostID:       "post",
		FormattedContent: drawdomain.FormattedContent(fortuneKeyword),
	})
	if !errors.Is(err, llm.ErrContentRejected) {
		t.Fatalf("expected rejected, got %v", err)
	}
	if result.Status != "rejected" {
		t.Fatalf("expected rejected status, got %s", result.Status)
	}
}

func TestFormatterValidateNil(t *testing.T) {
	f := &Formatter{}
	if _, err := f.Validate(context.Background(), nil); !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format, got %v", err)
	}
}

func TestFormatterValidateEmptyContent(t *testing.T) {
	f := &Formatter{}
	result, err := f.Validate(context.Background(), &llm.FormatResult{
		DarkPostID:       "post",
		FormattedContent: "   ",
	})
	if !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format for empty content, got %v", err)
	}
	if result.Status != "rejected" {
		t.Fatalf("expected rejected status")
	}
}

func TestExtractFirstText(t *testing.T) {
	text, err := extractFirstText(githubOpenAI.ChatCompletionResponse{
		Choices: []githubOpenAI.ChatCompletionChoice{{
			Message: githubOpenAI.ChatCompletionMessage{Content: " こんにちは "},
		}},
	})
	if err != nil || text != "こんにちは" {
		t.Fatalf("unexpected result: %q %v", text, err)
	}
}

func TestExtractFirstTextSkipsEmpty(t *testing.T) {
	text, err := extractFirstText(githubOpenAI.ChatCompletionResponse{
		Choices: []githubOpenAI.ChatCompletionChoice{
			{Message: githubOpenAI.ChatCompletionMessage{Content: "   "}},
			{Message: githubOpenAI.ChatCompletionMessage{Content: "text"}},
		},
	})
	if err != nil || text != "text" {
		t.Fatalf("expected to skip empty parts: %q %v", text, err)
	}
}

func TestExtractFirstTextAllEmpty(t *testing.T) {
	_, err := extractFirstText(githubOpenAI.ChatCompletionResponse{
		Choices: []githubOpenAI.ChatCompletionChoice{{
			Message: githubOpenAI.ChatCompletionMessage{Content: "   "},
		}},
	})
	if !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format when all choices empty")
	}
}

func TestExtractFirstTextError(t *testing.T) {
	_, err := extractFirstText(githubOpenAI.ChatCompletionResponse{})
	if !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format, got %v", err)
	}
}

func TestValidateFormatRequest(t *testing.T) {
	if err := validateFormatRequest(&llm.FormatRequest{DarkPostID: "id", DarkContent: "body"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateFormatRequest(&llm.FormatRequest{}); !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format, got %v", err)
	}
}

func TestValidateFormatRequestNil(t *testing.T) {
	if err := validateFormatRequest(nil); !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format for nil request")
	}
}

func TestValidateFormatRequestEmptyContent(t *testing.T) {
	if err := validateFormatRequest(&llm.FormatRequest{DarkPostID: "id", DarkContent: "   "}); !errors.Is(err, llm.ErrInvalidFormat) {
		t.Fatalf("expected invalid format for empty content")
	}
}

func TestShouldReject(t *testing.T) {
	if reason, ok := shouldReject(fortuneShort); !ok || !strings.Contains(reason, "短すぎます") {
		t.Fatalf("expected rejection for short text")
	}
	if reason, ok := shouldReject(fortuneLong); !ok || !strings.Contains(reason, "長すぎます") {
		t.Fatalf("expected rejection for long text")
	}
	if reason, ok := shouldReject(fortuneURL); !ok || !strings.Contains(reason, "URL") {
		t.Fatalf("expected rejection for URL")
	}
	if reason, ok := shouldReject(fortuneMissingPrefix); !ok || !strings.Contains(reason, "冒頭") {
		t.Fatalf("expected rejection for prefix")
	}
	if reason, ok := shouldReject(fortuneTwoSentences); !ok || !strings.Contains(reason, "3文") {
		t.Fatalf("expected rejection for sentence count")
	}
	if reason, ok := shouldReject(fortuneValid); ok || reason != "" {
		t.Fatalf("expected acceptance, got %v %v", ok, reason)
	}
}

func TestShouldRejectKeywords(t *testing.T) {
	if reason, ok := shouldReject(fortuneKeyword); !ok || !strings.Contains(reason, "kill") {
		t.Fatalf("expected keyword rejection, reason=%v", reason)
	}
}

func TestNewFormatterRequiresKey(t *testing.T) {
	if _, err := NewFormatter(" ", "model", ""); err == nil {
		t.Fatalf("expected error when key is missing")
	}
}

func TestNewFormatterDefaults(t *testing.T) {
	f, err := NewFormatter("dummy", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.model != config.DefaultOpenAIModel {
		t.Fatalf("expected default model, got %s", f.model)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close should succeed")
	}
}

func TestNewFormatterWithBaseURL(t *testing.T) {
	if _, err := NewFormatter("dummy", "gpt-test", "https://example.com"); err != nil {
		t.Fatalf("unexpected error with baseURL: %v", err)
	}
}

func TestBuildPromptTrims(t *testing.T) {
	got := buildPrompt(" こんにちは ")
	if !strings.Contains(got, "こんにちは") {
		t.Fatalf("prompt does not contain content: %s", got)
	}
}

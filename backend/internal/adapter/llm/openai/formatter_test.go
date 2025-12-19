package openai

import (
	"context"
	"errors"
	"strings"
	"testing"

	"backend/internal/port/llm"

	githubOpenAI "github.com/sashabaranov/go-openai"
)

type stubChatClient struct {
	resp githubOpenAI.ChatCompletionResponse
	err  error
}

func (s *stubChatClient) CreateChatCompletion(ctx context.Context, req githubOpenAI.ChatCompletionRequest) (githubOpenAI.ChatCompletionResponse, error) {
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

func TestFormatterValidate(t *testing.T) {
	f := &Formatter{}
	result, err := f.Validate(context.Background(), &llm.FormatResult{
		DarkPostID:       "post",
		FormattedContent: " 整形済み ",
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
		FormattedContent: "含む kill",
	})
	if !errors.Is(err, llm.ErrContentRejected) {
		t.Fatalf("expected rejected, got %v", err)
	}
	if result.Status != "rejected" {
		t.Fatalf("expected rejected status, got %s", result.Status)
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

func TestShouldReject(t *testing.T) {
	if reason, ok := shouldReject("visit http://example.com"); !ok || reason == "" {
		t.Fatalf("expected rejection for URL")
	}
	if reason, ok := shouldReject("clean text"); ok || reason != "" {
		t.Fatalf("expected acceptance, got %v %v", ok, reason)
	}
}

func TestBuildPromptTrims(t *testing.T) {
	got := buildPrompt(" こんにちは ")
	if !strings.Contains(got, "こんにちは") {
		t.Fatalf("prompt does not contain content: %s", got)
	}
}

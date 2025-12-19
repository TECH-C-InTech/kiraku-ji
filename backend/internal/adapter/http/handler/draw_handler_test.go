package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"

	"github.com/gin-gonic/gin"
)

func TestDrawHandler_GetRandomDraw(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		d := newVerifiedDraw(t, "post-success", "fortunes await")
		handler := NewDrawHandler(&stubFortuneUsecase{draw: d})
		router := NewRouter(handler, &PostHandler{})

		rec, body := performRequest(router)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d but got %d", http.StatusOK, rec.Code)
		}

		var got DrawResponse
		decodeBody(t, body, &got)

		want := DrawResponse{
			PostID: "post-success",
			Result: "fortunes await",
			Status: string(d.Status()),
		}

		if got != want {
			t.Fatalf("unexpected response: %+v", got)
		}
	})

	t.Run("draws depleted", func(t *testing.T) {
		handler := NewDrawHandler(&stubFortuneUsecase{err: drawdomain.ErrEmptyResult})
		router := NewRouter(handler, &PostHandler{})

		rec, body := performRequest(router)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d but got %d", http.StatusNotFound, rec.Code)
		}

		var got errorResponse
		decodeBody(t, body, &got)

		if got.Message != messageDrawsEmpty {
			t.Fatalf("expected message %q but got %q", messageDrawsEmpty, got.Message)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		handler := NewDrawHandler(&stubFortuneUsecase{err: errors.New("boom")})
		router := NewRouter(handler, &PostHandler{})

		rec, body := performRequest(router)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d but got %d", http.StatusInternalServerError, rec.Code)
		}

		var got errorResponse
		decodeBody(t, body, &got)

		if got.Message != messageInternalError {
			t.Fatalf("expected message %q but got %q", messageInternalError, got.Message)
		}
	})
}

type stubFortuneUsecase struct {
	draw *drawdomain.Draw
	err  error
}

func (s *stubFortuneUsecase) DrawFortune(ctx context.Context) (*drawdomain.Draw, error) {
	return s.draw, s.err
}

func newVerifiedDraw(t *testing.T, postID, result string) *drawdomain.Draw {
	t.Helper()
	d, err := drawdomain.New(post.DarkPostID(postID), drawdomain.FormattedContent(result))
	if err != nil {
		t.Fatalf("failed to create draw: %v", err)
	}
	d.MarkVerified()
	return d
}

func performRequest(router *gin.Engine) (*httptest.ResponseRecorder, *bytes.Buffer) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/draws/random", nil)
	router.ServeHTTP(rec, req)
	return rec, rec.Body
}

func decodeBody[T any](t *testing.T, body io.Reader, out *T) {
	t.Helper()
	if err := json.NewDecoder(body).Decode(out); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
}

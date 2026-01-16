package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	drawhandler "backend/internal/adapter/http/handler"
	"backend/internal/app"
)

type stubRouter struct {
	runErr error
	called bool
	args   []string
}

/**
 * 呼び出し状況を記録し、指定した結果を返す。
 */
func (s *stubRouter) Run(args ...string) error {
	s.called = true
	s.args = args
	return s.runErr
}

/**
 * 依存が正常な場合に起動処理が成功することを確認する。
 */
func TestRun_Success(t *testing.T) {
	origContainer := newContainer
	origRouter := newRouter
	origClose := closeContainer
	t.Cleanup(func() {
		newContainer = origContainer
		newRouter = origRouter
		closeContainer = origClose
	})

	expectedDraw := &drawhandler.DrawHandler{}
	expectedPost := &drawhandler.PostHandler{}
	container := &app.Container{DrawHandler: expectedDraw, PostHandler: expectedPost}
	newContainer = func(ctx context.Context) (*app.Container, error) {
		return container, nil
	}

	routerStub := &stubRouter{}
	var gotDraw *drawhandler.DrawHandler
	var gotPost *drawhandler.PostHandler
	newRouter = func(draw *drawhandler.DrawHandler, post *drawhandler.PostHandler) routerRunner {
		gotDraw = draw
		gotPost = post
		return routerStub
	}

	if err := run(context.Background()); err != nil {
		t.Fatalf("エラーなしを想定しましたが取得しました: %v", err)
	}
	if !routerStub.called {
		t.Fatalf("router.Run の呼び出しを想定しましたが未実行です")
	}
	if gotDraw != expectedDraw {
		t.Fatalf("draw handler の引き渡しが想定どおりではありません")
	}
	if gotPost != expectedPost {
		t.Fatalf("post handler の引き渡しが想定どおりではありません")
	}
}

/**
 * 依存初期化失敗がエラーとして返ることを確認する。
 */
func TestRun_NewContainerError(t *testing.T) {
	origContainer := newContainer
	origRouter := newRouter
	origClose := closeContainer
	t.Cleanup(func() {
		newContainer = origContainer
		newRouter = origRouter
		closeContainer = origClose
	})

	expectedErr := errors.New("コンテナ生成失敗")
	newContainer = func(ctx context.Context) (*app.Container, error) {
		return nil, expectedErr
	}
	newRouter = func(*drawhandler.DrawHandler, *drawhandler.PostHandler) routerRunner {
		t.Fatalf("router の生成は想定外です")
		return nil
	}

	err := run(context.Background())
	if err == nil {
		t.Fatalf("エラー発生を想定しましたが nil でした")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("期待したエラーが含まれていません: %v", err)
	}
	if !strings.Contains(err.Error(), "依存初期化失敗") {
		t.Fatalf("エラーメッセージに文脈が含まれていません: %v", err)
	}
}

/**
 * ルーター起動失敗がエラーとして返ることを確認する。
 */
func TestRun_RouterRunError(t *testing.T) {
	origContainer := newContainer
	origRouter := newRouter
	origClose := closeContainer
	t.Cleanup(func() {
		newContainer = origContainer
		newRouter = origRouter
		closeContainer = origClose
	})

	container := &app.Container{
		DrawHandler: &drawhandler.DrawHandler{},
		PostHandler: &drawhandler.PostHandler{},
	}
	newContainer = func(ctx context.Context) (*app.Container, error) {
		return container, nil
	}

	expectedErr := errors.New("起動失敗")
	newRouter = func(*drawhandler.DrawHandler, *drawhandler.PostHandler) routerRunner {
		return &stubRouter{runErr: expectedErr}
	}

	err := run(context.Background())
	if err == nil {
		t.Fatalf("エラー発生を想定しましたが nil でした")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("期待したエラーが含まれていません: %v", err)
	}
	if !strings.Contains(err.Error(), "サーバー起動失敗") {
		t.Fatalf("エラーメッセージに文脈が含まれていません: %v", err)
	}
}

/**
 * 依存終了の失敗があっても起動処理が成功することを確認する。
 */
func TestRun_CloseErrorIsLogged(t *testing.T) {
	origContainer := newContainer
	origRouter := newRouter
	origClose := closeContainer
	t.Cleanup(func() {
		newContainer = origContainer
		newRouter = origRouter
		closeContainer = origClose
	})

	container := &app.Container{
		DrawHandler: &drawhandler.DrawHandler{},
		PostHandler: &drawhandler.PostHandler{},
	}
	newContainer = func(ctx context.Context) (*app.Container, error) {
		return container, nil
	}
	newRouter = func(*drawhandler.DrawHandler, *drawhandler.PostHandler) routerRunner {
		return &stubRouter{}
	}

	closeCalled := false
	closeContainer = func(*app.Container) error {
		closeCalled = true
		return errors.New("終了失敗")
	}

	if err := run(context.Background()); err != nil {
		t.Fatalf("エラーなしを想定しましたが取得しました: %v", err)
	}
	if !closeCalled {
		t.Fatalf("close の呼び出しを想定しましたが未実行です")
	}
}

/**
 * 起動成功時に致命的ログが呼ばれないことを確認する。
 */
func TestMain_RunSuccessDoesNotCallFatalf(t *testing.T) {
	origRun := runFunc
	origFatalf := fatalf
	t.Cleanup(func() {
		runFunc = origRun
		fatalf = origFatalf
	})

	runFunc = func(ctx context.Context) error {
		return nil
	}
	called := false
	fatalf = func(format string, v ...any) {
		called = true
	}

	main()

	if called {
		t.Fatalf("fatalf の呼び出しは想定外です")
	}
}

/**
 * 起動失敗時に致命的ログが呼ばれることを確認する。
 */
func TestMain_RunErrorCallsFatalf(t *testing.T) {
	origRun := runFunc
	origFatalf := fatalf
	t.Cleanup(func() {
		runFunc = origRun
		fatalf = origFatalf
	})

	runFunc = func(ctx context.Context) error {
		return errors.New("起動失敗")
	}
	var gotMsg string
	fatalf = func(format string, v ...any) {
		gotMsg = fmt.Sprintf(format, v...)
	}

	main()

	if gotMsg == "" {
		t.Fatalf("fatalf の呼び出しを想定しましたが未実行です")
	}
	if !strings.Contains(gotMsg, "API起動失敗") {
		t.Fatalf("fatalf のメッセージに文脈が含まれていません: %s", gotMsg)
	}
}

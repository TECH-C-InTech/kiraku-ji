package main

import (
	"testing"

	drawhandler "backend/internal/adapter/http/handler"
	"backend/internal/app"
)

/**
 * デフォルトの依存生成とクローズが動作することを確認する。
 */
func TestMainDeps_DefaultDepsAreCallable(t *testing.T) {
	router := newRouter(&drawhandler.DrawHandler{}, &drawhandler.PostHandler{})
	if router == nil {
		t.Fatalf("router の生成に失敗しました")
	}

	if err := closeContainer(&app.Container{}); err != nil {
		t.Fatalf("container の close に失敗しました: %v", err)
	}
}

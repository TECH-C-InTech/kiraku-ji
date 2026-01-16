package main

import (
	"context"
	"fmt"
	"log"

	"backend/internal/config"
)

/**
 * API 実行に必要な前処理を済ませ、起動処理を呼び出す
 */
func main() {
	config.LoadDotEnv()

	if err := runFunc(context.Background()); err != nil {
		fatalf("API起動失敗: %v", err)
	}
}

/**
 * 依存を初期化し、HTTP サーバーを起動する
 */
func run(ctx context.Context) error {
	// 依存関係をまとめて初期化
	container, err := newContainer(ctx)
	if err != nil {
		return fmt.Errorf("依存初期化失敗: %w", err)
	}

	// 関数の終了時に依存リソースを閉じる
	defer func() {
		if closeErr := closeContainer(container); closeErr != nil {
			log.Printf("依存終了失敗: %v", closeErr)
		}
	}()

	// ルーティングを組み立てて、起動
	router := newRouter(container.DrawHandler, container.PostHandler)
	if err := router.Run(); err != nil {
		return fmt.Errorf("サーバー起動失敗: %w", err)
	}
	return nil
}

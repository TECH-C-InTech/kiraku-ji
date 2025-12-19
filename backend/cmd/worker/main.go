package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/app"
	"backend/internal/config"
	"backend/internal/port/queue"
)

/**
 * 起動時にワーカーの依存を整えて停止指示が来るまでループを回す。
 */
func main() {
	config.LoadDotEnv()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	container, err := app.NewWorkerContainer(ctx)
	if err != nil {
		log.Fatalf("failed to initialize worker: %v", err)
	}
	defer func() {
		if cerr := container.Close(); cerr != nil {
			log.Printf("worker shutdown error: %v", cerr)
		}
	}()

	log.Println("worker started (pending format)")
	runLoop(ctx, container)
}

/**
 * 取り出した投稿を順に整形し、終了指示や取り出し失敗を監視しながら回し続ける。
 */
func runLoop(ctx context.Context, container *app.WorkerContainer) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("worker shutting down: %v", ctx.Err())
			return
		default:
		}

		postID, err := container.JobQueue.DequeueFormat(ctx)
		if err != nil {
			// 中断やキュー停止はそのまま終了する
			if errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) ||
				errors.Is(err, queue.ErrQueueClosed) ||
				errors.Is(err, queue.ErrContextClosed) {
				return
			}
			// それ以外は短い待機後に再試行
			log.Printf("dequeue error: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if err := container.FormatPendingUsecase.Execute(ctx, string(postID)); err != nil {
			// LLM や投稿の整形問題はログに残して次のジョブへ
			log.Printf("format error (post=%s): %v", postID, err)
			continue
		}

		log.Printf("formatted post: %s", postID)
	}
}

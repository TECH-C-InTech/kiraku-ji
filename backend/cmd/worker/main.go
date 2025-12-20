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
	usecaseworker "backend/internal/usecase/worker"
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

		// ジョブを処理し、失敗内容ごとにログの粒度を変える
		if err := container.FormatPendingUsecase.Execute(ctx, string(postID)); err != nil {
			switch {
			// draw 保存に失敗したが再キュー済みのケース
			case errors.Is(err, usecaseworker.ErrDrawCreationFailed):
				log.Printf("draw creation failed (post=%s): %v (requeued)", postID, err)
				// 再キューやロールバック自体が失敗した致命的ケース
			case errors.Is(err, usecaseworker.ErrRequeueFailed):
				log.Printf("draw creation rollback failed (post=%s): %v", postID, err)
			default:
				// LLM や投稿の整形問題はログに残して次のジョブへ
				log.Printf("format error (post=%s): %v", postID, err)
			}
			continue
		}

		log.Printf("formatted post: %s", postID)
	}
}

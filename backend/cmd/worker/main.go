package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/app"
	"backend/internal/port/queue"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	container, err := app.NewWorkerContainer(ctx)
	if err != nil {
		log.Fatalf("failed to initialize worker: %v", err)
	}
	defer container.Close()

	log.Println("worker started (pending format)")
	runLoop(ctx, container)
}

func runLoop(ctx context.Context, container *app.WorkerContainer) {
	for {
		select {
		case <-ctx.Done():
			log.Println("worker shutting down")
			return
		default:
		}

		postID, err := container.JobQueue.DequeueFormat(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, queue.ErrQueueClosed) {
				return
			}
			log.Printf("dequeue error: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if err := container.FormatPendingUsecase.Execute(ctx, string(postID)); err != nil {
			log.Printf("format error (post=%s): %v", postID, err)
			continue
		}

		log.Printf("formatted post: %s", postID)
	}
}

package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	queueFirestore "backend/internal/adapter/queue/firestore"
	queueMemory "backend/internal/adapter/queue/memory"
	"backend/internal/port/queue"

	"cloud.google.com/go/firestore"
)

const (
	jobQueueBackendEnv       = "JOB_QUEUE_BACKEND"
	jobQueueBackendMemory    = "memory"
	jobQueueBackendFirestore = "firestore"
	defaultJobQueueBuffer    = 10
)

var (
	jobQueueFactory          = newJobQueue
	firestoreJobQueueFactory = func(client *firestore.Client) (queue.JobQueue, error) {
		return queueFirestore.NewFirestoreJobQueue(client)
	}
	memoryJobQueueConstructor = queueMemory.NewInMemoryJobQueue
)

var errFirestoreQueueRequiresClient = errors.New("job queue: firestore を選択する場合はクライアントが必要です")

/**
 * JOB_QUEUE_BACKEND の値に応じてメモリ or Firestore の整形キューを構築し、
 * サンプル投稿を投入すべきかどうかのフラグも返す。
 */
func newJobQueue(infra *Infra) (queue.JobQueue, bool, error) {
	mode := strings.TrimSpace(os.Getenv(jobQueueBackendEnv))
	// 明示が無い場合はメモリ実装を既定として扱う
	switch strings.ToLower(mode) {
	case "", jobQueueBackendMemory:
		return memoryJobQueueConstructor(defaultJobQueueBuffer), true, nil
	case jobQueueBackendFirestore:
		// Firestore を選ぶ場合は Infra 側でクライアントが初期化済みである必要がある
		if infra == nil || infra.Firestore() == nil {
			return nil, false, errFirestoreQueueRequiresClient
		}
		// Firestore 用の JobQueue 実体を生成し、メモリ時のような seed は不要
		jobQueue, err := firestoreJobQueueFactory(infra.Firestore())
		if err != nil {
			return nil, false, fmt.Errorf("new firestore job queue: %w", err)
		}
		return jobQueue, false, nil
	default:
		return nil, false, fmt.Errorf("job queue backend %q は未対応です", mode)
	}
}

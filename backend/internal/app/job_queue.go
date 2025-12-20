package app

import (
	"errors"
	"fmt"

	queueFirestore "backend/internal/adapter/queue/firestore"
	"backend/internal/port/queue"

	"cloud.google.com/go/firestore"
)

var (
	jobQueueFactory          = newJobQueue
	firestoreJobQueueFactory = func(client *firestore.Client) (queue.JobQueue, error) {
		return queueFirestore.NewFirestoreJobQueue(client)
	}
)

var errFirestoreQueueRequiresClient = errors.New("job queue: Firestore クライアントが初期化されていません")

/**
 * Firestore 固定の整形ジョブキューを構築する。
 */
func newJobQueue(infra *Infra) (queue.JobQueue, error) {
	if infra == nil || infra.Firestore() == nil {
		return nil, errFirestoreQueueRequiresClient
	}
	jobQueue, err := firestoreJobQueueFactory(infra.Firestore())
	if err != nil {
		return nil, fmt.Errorf("new firestore job queue: %w", err)
	}
	return jobQueue, nil
}

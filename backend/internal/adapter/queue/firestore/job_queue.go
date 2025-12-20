package firestore

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"backend/internal/domain/post"
	"backend/internal/port/queue"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	formatJobsCollection = "format_jobs"
	jobStatusPending     = "pending"
	pollInterval         = 200 * time.Millisecond
)

var (
	errMissingClient   = errors.New("firestorejobqueue: Firestore クライアントが指定されていません")
	errEmptyPostID     = errors.New("firestorejobqueue: 投稿 ID が指定されていません")
	errNoJobAvailable  = errors.New("firestorejobqueue: キューが空です")
	errDecodeJobFailed = errors.New("firestorejobqueue: ドキュメントの復元に失敗しました")
)

// Firestore に記録する整形ジョブ 1 件分の姿
type jobDocument struct {
	PostID string    `firestore:"post_id"`
	Status string    `firestore:"status"`
	Queued time.Time `firestore:"created_at"`
}

// Firestore を永続化に使う整形待ちキュー
type FirestoreJobQueue struct {
	client     *firestore.Client
	collection string
	closeOnce  sync.Once
	closedCh   chan struct{}
}

/**
 * Firestore 接続を受け取り、format_jobs を背後に使う整形キューを組み立てる。
 */
func NewFirestoreJobQueue(client *firestore.Client) (*FirestoreJobQueue, error) {
	if client == nil {
		return nil, errMissingClient
	}
	return &FirestoreJobQueue{
		client:     client,
		collection: formatJobsCollection,
		closedCh:   make(chan struct{}),
	}, nil
}

/**
 * 整形待ち投稿の ID を Firestore に書き込み、二重登録なら専用エラーを返す。
 */
func (q *FirestoreJobQueue) EnqueueFormat(ctx context.Context, id post.DarkPostID) error {
	if err := q.ensureReady(ctx); err != nil {
		return err
	}
	if id == "" {
		return errEmptyPostID
	}

	doc := q.client.Collection(q.collection).Doc(string(id))
	payload := map[string]any{
		"post_id":    string(id),
		"status":     jobStatusPending,
		"created_at": firestore.ServerTimestamp,
	}
	_, err := doc.Create(ctx, payload)
	if status.Code(err) == codes.AlreadyExists {
		return queue.ErrJobAlreadyScheduled
	}
	if err != nil {
		return translateContextError(err)
	}
	return nil
}

/**
 * Firestore 上で最も古い整形待ちを 1 件だけ取得し、見つかるまで待機を繰り返す。
 */
func (q *FirestoreJobQueue) DequeueFormat(ctx context.Context) (post.DarkPostID, error) {
	for {
		if err := q.ensureReady(ctx); err != nil {
			return "", err
		}
		id, err := q.dequeueOnce(ctx)
		if err == nil {
			return id, nil
		}
		// ジョブがまだ用意されていない場合は停止指示を監視しながら待機して再試行する
		if errors.Is(err, errNoJobAvailable) {
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("%w: %v", queue.ErrContextClosed, ctx.Err())
			case <-q.closedCh:
				return "", queue.ErrQueueClosed
			case <-time.After(pollInterval):
				continue
			}
		}
		return "", err
	}
}

/**
 * 以降の登録・取り出しを止めるため通知チャネルを閉じる。
 */
func (q *FirestoreJobQueue) Close() error {
	if q == nil {
		return nil
	}
	q.closeOnce.Do(func() {
		close(q.closedCh)
	})
	return nil
}

/**
 * 呼び出し側の中断や自身の停止状態を確認し、継続可否を判定する。
 */
func (q *FirestoreJobQueue) ensureReady(ctx context.Context) error {
	if q == nil {
		return queue.ErrQueueClosed
	}
	// 一度停止済みならこれ以上の登録や取り出しは受け付けない
	select {
	case <-q.closedCh:
		return queue.ErrQueueClosed
	default:
	}
	if ctx == nil {
		return fmt.Errorf("%w: context が nil です", queue.ErrContextClosed)
	}
	// 呼び出し側のコンテキストが終わっていないか先に確認する
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", queue.ErrContextClosed, ctx.Err())
	default:
		return nil
	}
}

/**
 * Firestore の format_jobs から一番古いジョブをトランザクションで取得し、その場で削除する。
 */
func (q *FirestoreJobQueue) dequeueOnce(ctx context.Context) (post.DarkPostID, error) {
	query := q.client.Collection(q.collection).OrderBy("created_at", firestore.Asc).Limit(1)
	var dequeued post.DarkPostID
	// トランザクションでドキュメント取得と削除をまとめ、複数ワーカーからの重複処理を避ける
	err := q.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		docs, err := tx.Documents(query).GetAll()
		if err != nil {
			return err
		}
		if len(docs) == 0 {
			return errNoJobAvailable
		}
		var job jobDocument
		if err := docs[0].DataTo(&job); err != nil {
			return fmt.Errorf("%w: %v", errDecodeJobFailed, err)
		}
		if job.PostID == "" {
			return fmt.Errorf("%w: post_id が空です", errDecodeJobFailed)
		}
		// 自身で削除に成功した時点でジョブ獲得とみなす
		if err := tx.Delete(docs[0].Ref); err != nil {
			if status.Code(err) == codes.NotFound {
				return errNoJobAvailable
			}
			return err
		}
		dequeued = post.DarkPostID(job.PostID)
		return nil
	}, firestore.MaxAttempts(5))
	// トランザクション結果をキュー用のエラーへ丸める
	if err != nil {
		if errors.Is(err, errNoJobAvailable) {
			return "", errNoJobAvailable
		}
		return "", translateContextError(fmt.Errorf("dequeue tx: %w", err))
	}
	return dequeued, nil
}

/**
 * コンテキスト関連のエラーを共通の ErrContextClosed にそろえて返す。
 */
func translateContextError(err error) error {
	if err == nil {
		return nil
	}
	// 呼び出し側の文脈が中断・期限切れなら queue.ErrContextClosed に読み替える
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %v", queue.ErrContextClosed, err)
	}
	return err
}

var _ queue.JobQueue = (*FirestoreJobQueue)(nil)

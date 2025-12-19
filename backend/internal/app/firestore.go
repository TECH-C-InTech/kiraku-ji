package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var errFirestoreProjectIDBlank = errors.New("firestore: GOOGLE_CLOUD_PROJECT is not set")

// FirestoreConfig は Firestore クライアント初期化に必要な設定を保持する。
type FirestoreConfig struct {
	ProjectID       string
	CredentialsFile string
	EmulatorHost    string
}

// Infra は外部リソースへの接続をまとめて保持する。
type Infra struct {
	firestoreClient *firestore.Client
}

// NewInfra は Firestore を含む外部依存を初期化して返す。
func NewInfra(ctx context.Context) (*Infra, error) {
	cfg, err := loadFirestoreConfigFromEnv()
	if err != nil {
		if errors.Is(err, errFirestoreProjectIDBlank) {
			return &Infra{}, nil
		}
		return nil, fmt.Errorf("load firestore config: %w", err)
	}

	client, err := newFirestoreClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Infra{
		firestoreClient: client,
	}, nil
}

// Firestore は Firestore クライアントを返す（設定されていない場合は nil）。
func (i *Infra) Firestore() *firestore.Client {
	if i == nil {
		return nil
	}
	return i.firestoreClient
}

// Close は保持しているリソースを順次クローズする。
func (i *Infra) Close() error {
	if i == nil || i.firestoreClient == nil {
		return nil
	}
	return i.firestoreClient.Close()
}

func loadFirestoreConfigFromEnv() (*FirestoreConfig, error) {
	projectID := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if projectID == "" {
		return nil, errFirestoreProjectIDBlank
	}

	return &FirestoreConfig{
		ProjectID:       projectID,
		CredentialsFile: strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")),
		EmulatorHost:    strings.TrimSpace(os.Getenv("FIRESTORE_EMULATOR_HOST")),
	}, nil
}

func newFirestoreClient(ctx context.Context, cfg *FirestoreConfig) (*firestore.Client, error) {
	opts := []option.ClientOption{}

	// エミュレータ利用時は認証不要なので Credentials は読み込まない。
	if cfg.CredentialsFile != "" && cfg.EmulatorHost == "" {
		creds, err := os.ReadFile(cfg.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("read credentials file: %w", err)
		}
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, creds))
	}

	client, err := firestore.NewClient(ctx, cfg.ProjectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("initialize firestore client: %w", err)
	}
	return client, nil
}

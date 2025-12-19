# プロジェクトのフォルダ構成

本プロジェクトのバックエンドは Go + クリーンアーキテクチャ（Hexagonal寄り） をベースに構成しています。
ハッカソン開発を前提に、理解しやすさ・拡張しやすさ・テストしやすさのバランスを重視しています。

## 設計方針

- API と Worker を分離（役割が違うため）

- ビジネスロジックは usecase に集約

- DB / LLM / HTTP は外部依存として分離

- LLM 処理は非同期（投稿時）

- テストは domain / usecase / adapter 単位で分けて記述

## ディレクトリ全体構成

```
├── cmd/
│   ├── api/
│   │   └── main.go          # HTTP API エントリポイント
│   └── worker/
│       └── main.go          # 非同期ワーカー（LLM整形）
│
├── internal/
│   ├── app/                 # 起動・DI設定
│   │   ├── api.go
│   │   └── worker.go
│   │
│   ├── domain/              # ドメイン層（純粋なルール）
│   │   ├── post/
│   │   │   ├── post.go
│   │   │   └── post_test.go
│   │   └── draw/
│   │       ├── draw.go
│   │       └── draw_test.go
│   │
│   ├── usecase/             # ユースケース層（アプリの中心）
│   │   ├── post/
│   │   │   ├── create_post.go
│   │   │   ├── get_post.go
│   │   │   └── *_test.go
│   │   ├── draw/
│   │   │   ├── draw_fortune.go
│   │   │   ├── get_draw.go
│   │   │   ├── resolve_draw.go
│   │   │   └── *_test.go
│   │   ├── stats/
│   │   │   └── get_stats.go
│   │   └── worker/
│   │       ├── format_pending.go   # LLM整形処理
│   │       └── *_test.go
│   │
│   ├── port/                # インターフェース定義
│   │   ├── repository/
│   │   │   ├── post_repository.go
│   │   │   └── draw_repository.go
│   │   ├── llm/
│   │   │   └── formatter.go
│   │   └── queue/
│   │       └── job_queue.go
│   │
│   └── adapter/             # 外部実装
│       ├── http/
│       │   ├── handler/
│       │   │   ├── post_handler.go
│       │   │   ├── draw_handler.go
│       │   │   └── *_test.go
│       │   └── router.go
│       │
│       ├── repository/
│       │   ├── firestore/
│       │   └── postgres/
│       │
│       ├── llm/
│       │   └── gemini/
│       │       └── formatter.go
│       │
│       └── queue/
│           └── cloudtasks/
│
├── configs/
├── migrations/              # Postgresを使う場合のみ
├── go.mod
└── README.md
```

## 各ディレクトリの役割

`cmd/`

アプリケーションの **エントリポイント**。

- `cmd/api`

    HTTP API サーバー起動用

- `cmd/worker`

    非同期ワーカー（pending → ready、LLM整形）用

API と Worker を分けることで、責務とスケールを明確にしています。

---

`internal/domain/`

**ドメイン層（最も内側）。**

- 投稿（闇）

- おみくじ（draw）

- ステータス遷移や制約

外部依存を一切持たず、純粋なビジネスルールのみを定義します。
ユニットテストの中心。

---

`internal/usecase/`

**アプリケーションの中核。**

- API / Worker から呼ばれる処理

- 「闇を投稿する」「おみくじを引く」「整形する」などの操作単位

Repository / LLM / Queue は interface（port）経由で利用します。
**最もテスト価値が高い層。**

---

`internal/port/`

外部依存との **境界インターフェース**。

- Repository（DB）

- LLM Formatter

- Job Queue

usecase はここだけに依存します。


---

`internal/adapter/`

外部サービスの **具体実装**。

- HTTP（handler / router）

- DB（Firestore / Postgres）

- LLM（Gemini）

- Queue（Cloud Tasks 等）

差し替え可能で、テストではモックに置き換えます。

---

## テスト方針

- `domain`：純粋なユニットテスト

- `usecase`：Repository / LLM をモックしてテスト

- `adapter/http`：httptest による API テスト

- DB 統合テスト

## 手動動作確認（/draws/random）

```
# 1. API サーバーを起動
cd backend
go run ./cmd/api

# 2. 別ターミナルからリクエスト
curl -i localhost:8080/draws/random
```

環境変数 `DRAW_REPOSITORY_MODE` によりリポジトリの挙動を切り替えられます。  

| モード | 起動例 | 期待されるレスポンス |
| --- | --- | --- |
| (空) / `default` | `DRAW_REPOSITORY_MODE=default go run ./cmd/api` | Verified が存在するため `200 OK` |
| `empty` | `DRAW_REPOSITORY_MODE=empty go run ./cmd/api` | Verified が無く `404 Not Found`（`message=no verified draws available`） |
| `error` | `DRAW_REPOSITORY_MODE=error go run ./cmd/api` | リポジトリ強制エラーで `500 Internal Server Error` |

## Firestore 設定

API / Worker から Firestore を利用する際は、`internal/app` が 1 度だけクライアントを生成し、各コンテナに共有されます。以下の環境変数を設定してください。

### 必須・任意の環境変数

| 変数名 | 役割 |
| --- | --- |
| `GOOGLE_CLOUD_PROJECT` | Firestore を利用する GCP プロジェクト ID（必須） |
| `GOOGLE_APPLICATION_CREDENTIALS` | 本番・Staging などで用いるサービスアカウント JSON のパス（エミュレータ利用時は不要） |
| `FIRESTORE_EMULATOR_HOST` | Firestore Emulator を利用する場合のホスト名（例: `localhost:8080`） |

`GOOGLE_CLOUD_PROJECT` が未設定の場合は Firestore クライアントは初期化されません（メモリ実装のみで動作）。

### ローカル開発（Firestore Emulator）

1. Firestore Emulator を起動  
   `gcloud beta emulators firestore start --host-port=localhost:8080`
2. 別ターミナルで環境変数をエクスポート  
   ```bash
   export GOOGLE_CLOUD_PROJECT=dark-fortune-dev
   export FIRESTORE_EMULATOR_HOST=localhost:8080
   ```
3. 必要に応じて API / Worker を起動  
   `go run ./cmd/api`

### 本番・リモート環境

1. Firestore を利用するプロジェクト ID を `GOOGLE_CLOUD_PROJECT` に設定。
2. 対象サービスアカウント JSON のパスを `GOOGLE_APPLICATION_CREDENTIALS` に設定。
3. `FIRESTORE_EMULATOR_HOST` は未設定（実サービス接続）。
4. `go run ./cmd/api` もしくはビルド済みバイナリを実行。

### コレクションスキーマ

| コレクション | 主キー | フィールド |
| --- | --- | --- |
| `posts/{post_id}` | `post_id` | `content` (string), `status` (`pending`/`ready`), `created_at`, `updated_at` |
| `draws/{post_id}` | `post_id` (Post と同じ ID) | `result` (string), `status` (`pending`/`verified`/`rejected`), `created_at` |

### 初期データ投入（シード）

`cmd/seed` が Firestore に posts/draws のサンプルデータを投入します。Verified な draw を含む状態が一度で作成されるため、API を Firestore に切り替えた後でもすぐに挙動を確認できます。

```
cd backend
export GOOGLE_CLOUD_PROJECT=your-project
# Firestore Emulator を使う場合は FIRESTORE_EMULATOR_HOST も設定
go run ./cmd/seed
```

エミュレータ利用時は `gcloud beta emulators firestore start --host-port=localhost:8080` を別ターミナルで起動してから実行してください。

---

## エラーコメント規約

- 日本語で記述すること

- 1文で理解できるように

- 例
```
errors.New("create_post: 入力が指定されていません")
```

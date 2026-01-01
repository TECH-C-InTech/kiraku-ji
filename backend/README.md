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

## 開発時の同時起動

API と Worker を同時に動かす場合は、別ターミナルで Worker を起動してください。

```
cd backend
go run ./cmd/worker
```

## Firestore 設定

API / Worker から Firestore を利用する際は、`internal/app` が 1 度だけクライアントを生成し、各コンテナに共有されます。以下の環境変数を設定してください。

### 必須・任意の環境変数

| 変数名 | 役割 |
| --- | --- |
| `GOOGLE_CLOUD_PROJECT` | Firestore を利用する GCP / Firebase プロジェクト ID（必須） |
| `GOOGLE_APPLICATION_CREDENTIALS` | Firestore へ接続するサービスアカウント JSON のパス（必須） |
| `FIRESTORE_EMULATOR_HOST` | Firestore Emulator を利用する場合のホスト名（Worker など開発用） |
| `GEMINI_API_KEY` | Gemini formatter を使用する際の API キー |
| `GEMINI_MODEL` | 利用する Gemini モデル名（未設定時は `gemini-2.5-flash`） |
| `OPENAI_API_KEY` | OpenAI formatter を使用する際の API キー |
| `OPENAI_MODEL` | 利用する OpenAI モデル名（未設定時は `gpt-4o-mini`） |
| `OPENAI_BASE_URL` | OpenAI 互換エンドポイントを使う場合の Base URL（通常は空で OK） |
| `LLM_PROVIDER` | `openai` / `gemini` を指定して使用する LLM を切り替え（未設定時は `openai`） |

`GOOGLE_CLOUD_PROJECT` / `GOOGLE_APPLICATION_CREDENTIALS` が未設定の場合、Infra の初期化が失敗し API / Worker は起動しません。Worker も API と同様に Firestore リポジトリ固定のため、必ず同じ環境変数を用意してください。JobQueue も Firestore 固定 (`format_jobs` コレクション) のため、切り替え用の環境変数は存在しません。

### API を Firestore へ接続する（エミュレータ非対応）

1. Firebase もしくは GCP で Firestore を有効化し、API から投稿を書き込むプロジェクト ID を決める。
2. 対象プロジェクトでサービスアカウント（Cloud Datastore User 権限以上）を作成し、JSON キーをダウンロードする。
3. JSON キーはリポジトリ外もしくは `.gitignore` に含まれるパス（例: `backend/service-account.json`）へ保存する。
4. `.env` またはシェルに以下を設定する。
   ```bash
   GOOGLE_CLOUD_PROJECT=your-project-id
   GOOGLE_APPLICATION_CREDENTIALS=/absolute/path/to/service-account.json
   ```
5. API を起動する。
   ```bash
   cd backend
   go run ./cmd/api
   ```
   必須環境変数が欠けている場合は起動時にエラーで停止する。
6. 別ターミナルから投稿を作り、Firestore `posts` コレクションに反映されることを確認する。
   ```bash
   curl -i -X POST http://localhost:8080/posts \
     -H "Content-Type: application/json" \
     -d '{"post_id":"post-123","content":"闇の投稿です"}'
   ```

> API は Firestore Emulator をサポートしていません。常に本番と同じ Firestore（サービスアカウント JSON 経由）へ接続してください。

ワーカーも同じ Firestore を共有します。Firestore 待ち受けが未設定のまま `go run ./cmd/worker` を起動した場合はエラーで即終了するため、API と同じく `GOOGLE_CLOUD_PROJECT` / `GOOGLE_APPLICATION_CREDENTIALS` を先に指定してください。

### コレクションスキーマ

| コレクション | 主キー | フィールド |
| --- | --- | --- |
| `posts/{post_id}` | `post_id` | `content` (string), `status` (`pending`/`ready`), `created_at`, `updated_at` |
| `draws/{post_id}` | `post_id` (Post と同じ ID) | `result` (string), `status` (`pending`/`verified`/`rejected`), `created_at` |


## ワーカー起動方法

`.env`（`backend/.env.example`）に LLM の API キー等を設定した上で、以下のコマンドで整形ワーカーを起動できます。

```
cd backend
go run ./cmd/worker
```

整形キューを監視し、`LLM_PROVIDER` で指定した LLM（`openai` が既定）で整形して公開準備へ進めます。`LLM_PROVIDER=gemini` を設定すると Gemini 実装に切り替わります。

Worker でも Firestore への書き込みが必須のため、API 起動時と同じ環境変数を設定してから実行してください。

```bash
cd backend
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_APPLICATION_CREDENTIALS=/absolute/path/to/service-account.json
go run ./cmd/worker
```

### 投稿〜整形までの動作確認

1. すべてのターミナルで Firestore 関連の環境変数を設定する。
   ```bash
   export GOOGLE_CLOUD_PROJECT=your-project-id
   export GOOGLE_APPLICATION_CREDENTIALS=$PWD/service-account.json
   export OPENAI_API_KEY=sk-xxx        # ダミーキーでも可
   export LLM_PROVIDER=openai          # Gemini を使う場合は gemini
   ```
2. ターミナル A で API を起動する。
   ```bash
   go run ./cmd/api
   ```
3. ターミナル B で Worker を起動する。起動後に `worker started (pending format)` が出力されれば待機状態。
   ```bash
   go run ./cmd/worker
   ```
4. ターミナル C から投稿 API を叩いてジョブを enqueue する。
   ```bash
   curl -i -X POST http://localhost:8080/posts \
     -H "Content-Type: application/json" \
     -d '{"post_id":"post-firestore-check","content":"Firestore への書き込み確認"}'
   ```
5. Firestore `format_jobs/post-firestore-check` が追加され、Worker のログに以下いずれかが出力されればジョブを取得できている。
   ```
   2025/12/20 12:34:56 formatted post: post-firestore-check
   # もしくは LLM の鍵がダミーの場合
   2025/12/20 12:34:56 format error (post=post-firestore-check): format_pending: 整形サービスに接続できません
   ```
   LLM の鍵が有効なら `posts/post-firestore-check` の `status` が `ready` へ更新され、`format_jobs` からドキュメントが削除される。

### LLM ごとの設定例

1. **OpenAI を使う場合（既定）**
   ```bash
   cd backend
   export OPENAI_API_KEY=sk-xxx
   export OPENAI_MODEL=gpt-4o-mini # 省略可
   export LLM_PROVIDER=openai
   go run ./cmd/worker
   ```

2. **Gemini を使う場合**
   ```bash
   cd backend
   export GEMINI_API_KEY=xxxx
   export GEMINI_MODEL=gemini-2.5-flash # 省略可
   export LLM_PROVIDER=gemini
   go run ./cmd/worker
   ```
### 投稿→整形→draw 生成フロー

投稿 API から整形ワーカー、draw 公開までの処理を図にしたメモを `docs/draw_flow.md` に置いています。  
投稿保存→`format_jobs` キュー→FormatPendingUsecase→`draws` の流れや、手動検証手順の参考にしてください。

---

## エラーコメント規約

- 日本語で記述すること

- 1文で理解できるように

- 例
```
errors.New("create_post: 入力が指定されていません")
```

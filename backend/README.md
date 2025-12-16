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

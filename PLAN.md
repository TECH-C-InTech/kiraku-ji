# Cloud Run + Terraform + GitHub Actions（APIのみ）プラン

## 目的
- バックエンドAPI（`backend/cmd/api`）を **Cloud Run** にデプロイできるようにする
- GCPリソースは **Terraform** でIaC化する
- デプロイは **GitHub ActionsでCD**（OIDC / Workload Identity Federationで鍵なし）
- Worker / キュー（Cloud Tasks等）は **後回し**（動く最小を先に作る）

## 今回の方針（決定事項）
- **aqua でツール管理**（terraform, gh, tflint など）
- Terraform の Google Provider で認証するため、gcloud CLI は不要

## 現状（リポジトリの前提）
- APIエントリポイント: `backend/cmd/api/main.go`
- ルーティング: `backend/internal/adapter/http/handler/router.go`
  - 現状は `GET /draws/random` のみ
- `router.Run()` は引数なしで起動するためデフォルトで **:8080**（Cloud Runの`PORT=8080`と噛み合う）

## 事前に手動でやること（Terraformの前に最低限）
- GCPプロジェクト作成（Billing有効化）
  - リージョンはひとまず `asia-northeast1` 推奨（変数化して後で変えられる）

## 実装タスク（順序）
### ✅ 1) devcontainer にツール環境を整備（完了）
- aqua-installer feature を追加
- aqua で terraform, gh, tflint を管理

### 2) APIをCloud Runで動かすためのコンテナ化
- `backend/Dockerfile` を追加（multi-stageで `go build ./cmd/api`）
- Cloud Runで起動するバイナリは `cmd/api` のものだけに絞る

### 3) CORS（Vercelフロントから叩くため）
- `backend/internal/adapter/http/handler/router.go` にCORSミドルウェアを追加
- 環境変数 `CORS_ALLOW_ORIGINS`（カンマ区切り）で許可Originを制御できるようにする
  - 初期は `*` でも良い（学習優先）。本番でVercelのOriginに絞る

### 4) Terraform（API単体に必要なGCPリソース）
ディレクトリ例: `infra/`（中身はTerraform）

作るリソース（最小）
- Artifact Registry（Docker repo）
- Cloud Run service（API）
- Cloud Run実行用Service Account
- Cloud Run Invoker（いったん `allUsers` で公開）
- 必要APIの有効化（run / artifactregistry / iam / serviceusage 等）

変数（必須）
- `project_id`
- `region`（default: `asia-northeast1`）
- `service_name`
- `allow_unauthenticated`（default: true）

### 5) Terraform（GitHub Actions OIDCのためのWIF）
- Workload Identity Pool / Provider（GitHub OIDC）
- デプロイ用Service Account
- 最小権限のIAM付与
  - run.admin
  - artifactregistry.writer
  - iam.serviceAccountUser（Cloud Run実行SAを使えるように）

### 6) GitHub Actions（CD）
- mainへのpushで実行
- OIDCでGCP認証 → Docker build → Artifact Registry push → Cloud Run deploy
- deploy後に `GET /draws/random` を叩くスモークを入れる

## ローカルでのTerraform実行（最初の1回だけ）
devcontainer内で:
```bash
# Terraform の Google Provider で認証
# （Application Default Credentials を使用）
terraform init
terraform apply
```

## テスト戦略（やりすぎず、でも学びは出す）
- Terraform: `terraform fmt -check` / `terraform validate` + `terraform test`（checkブロック）をPRで回す
- デプロイ: GitHub Actionsでスモーク（`/draws/random` が 200 or 404 の想定で返る）

## 後回しにする拡張（次フェーズ）
- Worker（Cloud Run service or Jobs）とCloud Tasksを追加して非同期化
- Firestore本番利用（`GOOGLE_CLOUD_PROJECT` 等の運用整備）
- 認証/レート制限/監視（必要になった時点で）

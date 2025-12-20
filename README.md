# 技術スタック概要

本プロジェクトで採用している主要な技術要素をまとめています。詳細な設計やセットアップ手順は `backend/README.md` を参照してください。

## バックエンド
- 言語: Go 1.25 系
- フレームワーク/主要ライブラリ: Gin、Google Cloud Firestore クライアント、Google Generative AI SDK、OpenAI クライアント、gRPC
- アーキテクチャ: クリーンアーキテクチャ（Hexagonal 寄り）。API サーバーと非同期 Worker を分離し、usecase 層にビジネスロジックを集約し、外部依存は adapter/port で分離
- データベース: Google Cloud Firestore（本番環境に接続。Emulator は未サポート）
- デプロイ: Cloud Run でバックエンドプロセスをホスティング

## フロントエンド
- フレームワーク: Next.js 16（React 19）
- 言語/ツール: TypeScript、Tailwind CSS 4 系、Biome、babel-plugin-react-compiler
- アーキテクチャ方針: UI/Application/Domain/Infrastructure の 4 層で責務を分離し、依存方向を内向きに固定したクリーンアーキテクチャ志向の構成
- デプロイ: Vercel でホスティングしており、main へのマージ後に自動デプロイされる（GitHub Actions 側は lint/build の検証のみ実施）
- CI: `.github/workflows/frontend-ci.yml` で PR ごとに lint・format・build をチェック

## 開発環境 / ツール
- Dev Container: `.devcontainer/devcontainer.json` を利用。Node.js 24 / Go 1.25 / Docker outside of Docker を含むベースイメージに `aqua` のインストーラーを組み込み、コンテナ生成時に `aqua i` とグローバル CLI のセットアップを自動実行
- aqua: CLI のバージョンを一元管理するツールチェーン。`aqua/aqua.yaml` と `aqua/imports/*.yaml` に Terraform / TFLint などのピン止め設定をまとめ、`aqua i` で再現性のあるインストールを行う。開発端末や CI で同じバージョンの `terraform` や `tflint`、`buf` などを取得し、PATH に自動連携するため、追加で手動インストールする必要がない
- Docker: Cloud Run へデプロイする API サーバー用のコンテナは `backend/Dockerfile` でマルチステージビルドし、非 root ユーザー実行と PORT 環境変数対応を含めて最小構成にしている（ローカルビルド時は Dev Container で Docker を直接利用可能）

## 開発フロー
- バックエンド: `aqua i` で CLI を準備し、`go run ./cmd/api` または `go run ./cmd/worker` で起動（依存は Go Modules で解決）
- フロントエンド: `pnpm install` の後に `pnpm dev` でローカル起動（http://localhost:3000）

## インフラ / IaC
- Terraform と TFLint を `aqua/imports/terraform.yaml` / `aqua/imports/tflint.yaml` でバージョン固定し、Cloud Run や Firestore など GCP リソースを IaC で管理するためのツールチェーンを整備（現時点では具体的な Terraform 定義は未同梱）

## 関連ドキュメント
- バックエンド詳細: [`backend/README.md`](backend/README.md)

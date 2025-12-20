# フロントエンド README

Next.js（App Router）+ TypeScript のフロントエンドです。ディレクトリ構成は以下の記事を参考に整理しています。  
参考: https://zenn.dev/yamu_official/articles/70f59488e8415d

## 起動手順

```bash
pnpm install
pnpm dev
```

ブラウザで `http://localhost:3000` を開くと画面を確認できます。

## ディレクトリ構成

現状は `src/app` と `src/lib` のみで運用しています。必要になったら以下のディレクトリを追加する方針です。

```
frontend/
├── public/                  # 画像などの静的アセット
└── src/
    ├── app/                 # ルーティング・ページ・レイアウト・サーバーアクション
    ├── components/          # 再利用する UI 部品（必要に応じて追加）
    ├── hooks/               # カスタムフック（use~ 系、必要に応じて追加）
    ├── lib/                 # API クライアントや外部ライブラリ設定
    ├── utils/               # use~ 以外の汎用処理（必要に応じて追加）
    ├── styles/              # グローバル CSS やテーマ（必要に応じて追加）
    ├── types/               # 型定義（必要に応じて追加）
    ├── stores/              # グローバルステート（必要に応じて追加）
    └── constants/           # 定数（必要に応じて追加）
```

## 各ディレクトリの役割

- `app`: 画面ルーティング、レイアウト、ページ、サーバーアクションを配置する。
- `components`: 汎用 UI 部品やレイアウト部品を置く（Atomic Design でも UI/Layout 分離でも可）。
- `hooks`: UI から切り出した再利用ロジック（`use~`）をまとめる。
- `lib`: API クライアントやライブラリ初期化など、外部依存の薄いラッパーを置く。
- `utils`: `use~` ではない汎用関数を集約する。
- `styles`: 全体に効かせる CSS やテーマ関連を置く（`app/globals.css` でも可）。
- `types`: UI / API などの型定義をまとめる。
- `stores`: 画面全体で共有する状態を管理する。
- `constants`: 変更が少ない固定値を置く。
- `public`: 静的ファイル（画像・フォントなど）を置く。

## テスト指針

- UI: React Testing Library 等によるコンポーネント統合テストを基本とし、主要フローは E2E（例: Playwright）でカバーする。
- lib: API クライアントはモックサーバーを使った統合テストで契約とハンドリングを確認する。

# codex 開発ガイドライン

このプロジェクトでcodexを使用する際のルールです。

## 必読ドキュメントと共通言語

- 作業を開始する前に、必ずリポジトリ直下の `README.md` および `backend/README.md` を読み、前提と設計方針を把握してください。Codex も例外なく遵守すること。
- Codex を含むすべてのエージェントは、日本語で質問・回答・コメントを行うこと。英語や他言語での応答は禁止です。


## ブランチ命名規則

ブランチ名には必ずprefixを付けてください。

### 使用可能なprefix
- `feat` - 新機能追加
- `fix` - バグ修正
- `docs` - ドキュメント
- `refactor` - リファクタリング（機能変更なし）
- `test` - テスト追加・修正
- `ci` - CI関連
- `chore` - その他（雑務、ツール設定など）

### 例
```
feat/user-authentication
fix/login-button-crash
update/error-handling
refactor/api-client
docs/readme-update
chore/eslint-setup
```

## コミットメッセージ

コミットメッセージは**簡潔**に書いてください。必ずprefixを付けてください。


### 形式
```
prefix: 変更内容の簡潔な説明
```

### 例
```
feat: ユーザー認証機能を追加
fix: ログインボタンのクラッシュを修正
update: エラーハンドリングを改善
refacotor: APIクライアントのコードを整理
docs: READMEにセットアップ手順を追加
chore: ESLintの設定を追加
```

## Pull Request（PR）

PRのタイトルとボディは**日本語**で**シンプル**に書いてください。

### タイトル形式
コミットメッセージと同じ形式で構いません。
```
feat: ユーザー認証機能
```

### ボディ
- 何を変更したか
- なぜ変更したか（必要に応じて）
- 動作確認方法（必要に応じて）

簡潔に、わかりやすく書いてください。過度に詳しく書く必要はありません。

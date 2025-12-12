# Agents 開発ガイドライン

このプロジェクトでAgentsを使用する際のルールです。

## ブランチ命名規則

ブランチ名には必ずprefixを付けてください。

### 使用可能なprefix
- `add` - 新機能追加
- `fix` - バグ修正
- `update` - 既存機能の更新
- `refac` - リファクタリング
- `docs` - ドキュメント
- `chore` - その他（設定、ビルド、雑務など）

### 例
```
add/user-authentication
fix/login-button-crash
update/error-handling
refac/api-client
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
add: ユーザー認証機能を追加
fix: ログインボタンのクラッシュを修正
update: エラーハンドリングを改善
refac: APIクライアントのコードを整理
docs: READMEにセットアップ手順を追加
chore: ESLintの設定を追加
```

### Agentsを使用したコミット

Agentsを使用してコミットを作成する場合は、必ず`Co-Authored-By`を含めてください。

```
add: ユーザー認証機能を追加

Co-authored-by: Codex (AI assistant) [codex-bot@noreply.example.com](mailto:codex-bot@noreply.example.com)
```

これにより、AIが作成したコミットであることが明確になります。

## Pull Request（PR）

PRのタイトルとボディは**日本語**で**シンプル**に書いてください。

### タイトル形式
コミットメッセージと同じ形式で構いません。
```
add: ユーザー認証機能
```

### ボディ
- 何を変更したか
- なぜ変更したか（必要に応じて）
- 動作確認方法（必要に応じて）

簡潔に、わかりやすく書いてください。過度に詳しく書く必要はありません。

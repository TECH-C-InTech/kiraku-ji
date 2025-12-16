# Claude Code 開発ガイドライン

このプロジェクトでClaude Codeを使用する際のルールです。

## ブランチ命名規則

ブランチ名には必ずprefixを付けてください（prefixはConventional Commitsのtypeに準拠します）。

### 使用可能なprefix
- `feat` - 新機能追加
- `fix` - バグ修正
- `docs` - ドキュメント
- `refactor` - リファクタリング（機能変更なし）
- `test` - テスト追加・修正
- `perf` - パフォーマンス改善
- `build` - ビルド関連（依存関係、ビルドツールなど）
- `ci` - CI関連
- `chore` - その他（雑務、ツール設定など）

### 例
```
feat/user-authentication
fix/login-button-crash
refactor/api-client
docs/readme-update
chore/eslint-setup
```

## コミットメッセージ

コミットメッセージは **Conventional Commits 1.0.0** に準拠します。短く、意図が分かるように書いてください。

### 形式
```
<type>[optional scope][optional !]: <description>

[optional body]

[optional footer(s)]
```

### ルール
- `type` は上のprefix一覧から選ぶ（基本は `feat` / `fix` / `refactor` / `docs` / `test` / `chore`）
- `scope` は任意（例: `feat(api): ...` / `fix(frontend): ...`）
- 破壊的変更は `!` を付ける（例: `feat(api)!: ...`）
- 破壊的変更の説明を明確にしたい場合は footer に `BREAKING CHANGE: ...` を書く

### 例
```
feat: ユーザー認証機能を追加
fix: ログインボタンのクラッシュを修正
refactor: APIクライアントのコードを整理
docs: READMEにセットアップ手順を追加
chore: ESLintの設定を追加

feat(api): 認証トークンの検証を追加
fix(frontend): ログイン画面での例外を修正

feat(api)!: 認証エンドポイントを刷新

BREAKING CHANGE: `/v1/auth/login` を廃止し `/v2/auth/login` に変更
```

### Claude Codeを使用したコミット

Claude Codeを使用してコミットを作成する場合は、footer（trailer）として `Co-authored-by:` を含めてください。

```
feat: ユーザー認証機能を追加

Co-authored-by: Claude Sonnet 4.5 <noreply@anthropic.com>
```

これにより、AIが作成したコミットであることが明確になります。

## Pull Request（PR）

PRのタイトルとボディは**日本語**で**シンプル**に書いてください。

### タイトル形式
コミットメッセージと同じ形式（Conventional Commits）にしてください。
```
feat: ユーザー認証機能
```

### ボディ
- 何を変更したか
- なぜ変更したか（必要に応じて）
- 動作確認方法（必要に応じて）

簡潔に、わかりやすく書いてください。過度に詳しく書く必要はありません。

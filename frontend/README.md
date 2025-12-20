# フロントエンド README

Next.js（React）をベースにしたフロントエンドです。バックエンドのクリーンアーキテクチャ（Hexagonal 寄り）に合わせ、責務の分離・拡張しやすさ・テスト容易性を重視しています。

## 起動手順

```bash
pnpm install
pnpm dev
```

ブラウザで `http://localhost:3000` を開くと画面を確認できます。

## 設計方針（バックエンドと整合）

- ビジネスロジックは Application（ユースケース）に集約し、UI は入出力と表示に専念させる。
- HTTP やブラウザ API などの外部依存は Infrastructure に寄せ、Domain / Application は外部を知らない。
- 依存方向は内向き一方向（UI → Application → Domain）。外側は内側を知るが、逆依存は持たない。
- 原則は関心の分離。副作用を押し出し、テストしやすい形を保つ。

## レイヤ構成と責務

- UI: ページ・コンポーネント・スタイル。ユーザー操作を受け、Application のユースケースを呼び出す。
- Application: 画面の振る舞い・状態管理・フォーム調停。Domain のモデルを使い、Infrastructure 経由で外部と通信する。
- Domain: 純粋なビジネスルール。エンティティや値オブジェクトをここに置き、副作用なしで完結させる。
- Infrastructure: API クライアント、ブラウザ依存処理などの具体実装。Application/Domain のインターフェースに従う。

## テスト指針

- Domain: 副作用を排除したユニットテストでロジックを検証する。
- Application: ユースケース単位でモックを用い、状態遷移や分岐を確認する。
- UI: React Testing Library 等によるコンポーネント統合テストを基本とし、主要フローは E2E（例: Playwright）でカバーする。
- Infrastructure: API クライアントはモックサーバーを使った統合テストで契約とハンドリングを確認する。

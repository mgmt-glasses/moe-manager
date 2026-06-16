# moe-manager

秘書キャラ AI による、タスク管理・娯楽時間管理・AIチャットを統合したバックエンド API。

## 必要環境

- Go 1.22+
- PostgreSQL 16+
- Docker（ローカル DB 起動用）

## セットアップ

```bash
# 1. リポジトリをクローン
git clone <repo-url>
cd moe-manager

# 2. 環境変数を用意
cp .env.example .env
# .env を編集して DATABASE_URL などを設定

# 3. 依存パッケージを取得
go mod download
```

## ローカル起動

```bash
# PostgreSQL を Docker で起動
docker compose up -d db

# API サーバーを起動
make run
```

## 主要コマンド

```bash
make build    # ビルド
make test     # テスト
make vet      # go vet
make verify   # vet + test + build（CI と同等）
make run      # API サーバー起動
```

## ディレクトリ構成

```
cmd/api/                    # HTTP サーバー起動・ルーティング・依存性注入
internal/<domain>/          # model・service・interface・handler
internal/<domain>/adapter/  # PostgreSQL・LLM・TTS 等の実装
migrations/                 # PostgreSQL マイグレーション
docs/                       # 設計方針・API 仕様・Issue ドキュメント
```

## ドキュメント

- [開発方針](docs/development-policy.md)
- [アーキテクチャ方針](docs/architecture-guidelines.md)
- [実装ガイド](docs/implementation-guide.md)
- [API 方針](docs/api-guidelines.md)
- [DB 方針](docs/database-guidelines.md)
- [Issue ドキュメント](docs/Issue/README.md)
- [ADR](docs/adr/README.md)

## コントリビュート

[CONTRIBUTING.md](CONTRIBUTING.md) を参照。

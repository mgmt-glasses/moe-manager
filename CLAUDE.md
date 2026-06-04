# CLAUDE.md

このファイルは、リポジトリ内のコードを操作する際のガイダンスです。

## 現行方針

- MVP バックエンドの使用言語は Go。
- PostgreSQL を全ドメインで共有する。
- モジュラーモノリスとヘキサゴナルアーキテクチャを採用する。
- 現行方針は `docs/architecture-guidelines.md`、`docs/implementation-guide.md`、`docs/development-policy.md` を正とする。
- `packages/` と `apps/gateway/` の Python 実装はプロトタイプ・旧実装として参照してよいが、新規機能を追加しない。
- `docs/detail/ARCHITECTURE.md` と `docs/detail/Rule.md` の Python / FastAPI 構成は旧設計として扱う。

## Go 構成

```txt
cmd/api/                    # HTTP サーバー起動、ルーティング、依存性注入
internal/<domain>/          # model、service、interface、handler
internal/<domain>/adapter/  # PostgreSQL、LLM、TTS 等の実装
migrations/                 # PostgreSQL マイグレーション
```

## 依存ルール

- model / service は HTTP、DB driver、外部 API SDK、他ドメインの実装型に依存しない。
- 外部依存は利用側パッケージで定義した interface の背後に置く。
- 複数ドメインをまたぐユースケースは、その処理を所有する `internal/<domain>` に置く。
- `cmd/api` はルーティングと依存性注入に徹し、ビジネスロジックを持たない。
- Python 製 TTS を利用する場合は別サービスとして起動し、Go adapter から HTTP 等で接続する。

## コマンド

Go 基盤作成後は以下を基本コマンドとする。

```bash
go mod download
go test ./...
go run ./cmd/api
```

## 開発ルール

- ブランチは `develop` から切り、PR は `develop` に出す。
- コミットは Conventional Commits に準拠する。
- 原則 1 PR = 1 目的とし、変更範囲を明確にする。
- 実装ルールや設計判断を変更した場合は関連ドキュメントも更新する。
- ユーザが作成した未関連の変更を削除・巻き戻ししない。

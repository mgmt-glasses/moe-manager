# Architecture Decision Records (ADR)

設計判断の記録。「なぜそうしたか」を残すことで、将来の変更判断を助ける。

## 形式

ファイル名: `NNNN-<short-title>.md`（例: `0001-go-hexagonal-architecture.md`）

```markdown
# ADR-NNNN: タイトル

## ステータス

Accepted / Deprecated / Superseded by ADR-XXXX

## 背景

なぜこの判断が必要だったか。

## 判断

何を決めたか。

## 理由

選んだ理由、却下した代替案。

## 結果

この判断がもたらす影響。
```

## 一覧

| No. | タイトル | ステータス |
|-----|---------|----------|
| [0001](0001-go-migration.md) | Go 言語への移行 | Accepted |
| [0002](0002-postgresql-multi-tenant.md) | PostgreSQL 共有とマルチテナント | Accepted |
| [0003](0003-hexagonal-architecture.md) | ヘキサゴナルアーキテクチャ | Accepted |
| [0004](0004-screentime-aggregation.md) | スクリーンタイム集計 | Accepted |
| [0005](0005-document-restructure.md) | ドキュメント再編 | Accepted |
| [0006](0006-llm-context-caching.md) | LLM コンテキストキャッシュ | Accepted |
| [0007](0007-tts-decoupling.md) | TTS 分離 | Accepted |
| [0008](0008-langextract-integration.md) | LangExtract 統合 | Accepted |
| [0009](0009-deployment-cloud-run-cloud-sql.md) | デプロイ構成 Cloud Run + Cloud SQL | Accepted |

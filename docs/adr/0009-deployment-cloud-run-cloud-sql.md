# ADR-0009: デプロイ構成を Cloud Run + Cloud SQL にする

- ステータス: Accepted
- 日付: 2026-06-20
- 関連: docs/deployment.md, docs/database-guidelines.md, ADR-0007

## コンテキスト

Go API を本番・検証環境で稼働させ、フロントエンドから接続できる状態にする必要がある。
構成要素ごとに置き場所を決める。

- DB は `docs/database-guidelines.md` で「将来的にマネージド PostgreSQL への移行を想定」「接続先 URL の変更で移行できる」と方針化されている。
- TTS は ADR-0007 で別サービスに分離済み。GPU を要し、API と同じ環境では動かせない。
- LLM は Vertex AI（Gemini）を利用し、GCP プロジェクト `moe-manager` で aiplatform API が有効。
- `cmd/api` は起動時に DB 接続とマイグレーション適用を行う設計（`docs/database-guidelines.md` のマイグレーション方針）。

## 決定

Go API を **Cloud Run** にデプロイする。

- **DB**: Cloud SQL（マネージド PostgreSQL）に接続する。Cloud Run からは `--add-cloudsql-instances` で接続し、`DATABASE_URL` は Secret Manager で管理する。
- **ビルド**: Dockerfile から `docker build` でイメージを作り、Artifact Registry に push する。GitHub Actions の `deploy.yml` がこれを担う。
- **CI/CD**: GitHub Actions + Workload Identity Federation（長期キーなし）。
- **デプロイ対象ブランチ**: `develop`（デフォルトブランチかつ検証/ステージング環境）。push で自動デプロイし、`workflow_dispatch` でも手動実行できる。本番（`main`）デプロイは認証導入とあわせて別途整備する。
- **TTS**: 当面デプロイ対象外。`TTS_SERVICE_URL` 未設定でも API は起動し、音声機能のみ縮退する。

## 理由

- Cloud SQL は `docs/architecture-guidelines.md` で想定する PostgreSQL 共有方針に直接合致し、設計どおりの構成で稼働できる。
- model/service は repository interface にのみ依存するため、DB 実装の差し替えは接続先変更で済む。将来別のマネージド PostgreSQL に移す場合も `DATABASE_URL` 変更で対応できる。
- Workload Identity Federation により長期鍵を持たずに GitHub Actions から GCP へ認証でき、Artifact Registry へのビルド・push と Cloud Run デプロイを CI で一貫して実行できる。
- 検討した代替案と却下理由:
  - **Supabase（マネージド PostgreSQL）**: 無料枠で MVP 検証ができるが、Cloud SQL のほうが GCP 内で Vertex AI・Cloud Run と同一プロジェクトに閉じられ、IAM・ネットワーク・課金を一元管理できる。本番想定の構成と差が出る運用検証を避けるため Cloud SQL を採用する。コスト最適化が必要になった時点で再検討する。
  - **`gcloud run deploy --source`（Cloud Build）**: docker daemon を要さない利点があるが、Artifact Registry への明示的な push を CI で管理するほうがイメージタグ（`github.sha`）の追跡とロールバックが容易なため、Dockerfile からの `docker build`/push を採用する。
  - **DB 接続失敗時も起動を継続する一時改修**: 起動時マイグレーションを前提とする DB 方針を崩すため却下。

## 影響

- 起動時に golang-migrate（`embed.FS` 同梱）が Cloud SQL にスキーマを適用する。マイグレーションが失敗するとデプロイは失敗扱いになる（設計どおり）。
- `DATABASE_URL` は Cloud SQL の Unix socket 接続（`host=/cloudsql/PROJECT:REGION:INSTANCE`）を用いる。`deploy.yml` で `--add-cloudsql-instances` を指定する。
- `--allow-unauthenticated` は検証用に URL を公開する暫定措置。本番前に認証を導入する（残課題）。
- 音声ファイルは Cloud Run のエフェメラルディスクに保存されるため、スケール・再起動で失われる。永続化が必要になった時点で `AudioStorage` の外部ストレージ実装（GCS 等）を追加する。

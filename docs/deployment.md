# デプロイ手順（GCP Cloud Run）

Go API を GCP Cloud Run にデプロイする手順をまとめます。

## 構成概要

| コンポーネント | サービス |
|---------------|---------|
| Go API | Cloud Run |
| PostgreSQL | Cloud SQL (PostgreSQL) |
| LLM | Vertex AI (Gemini) |
| TTS | 別サービス（HTTP 接続、`TTS_SERVICE_URL` で指定） |
| コンテナレジストリ | Artifact Registry |
| CI 認証 | Workload Identity Federation (WIF) |

設計方針どおり、PostgreSQL / LLM / TTS は Cloud Run の単一デプロイには含めず外部サービスとして接続します。

## 事前準備（初回のみ）

### 1. GCP リソース作成

```bash
PROJECT_ID=moe-manager
REGION=us-central1

# Artifact Registry リポジトリ
gcloud artifacts repositories create moe-manager \
  --repository-format=docker --location=$REGION

# Cloud SQL (PostgreSQL) インスタンス
gcloud sql instances create moe-manager-db \
  --database-version=POSTGRES_16 --tier=db-f1-micro --region=$REGION
gcloud sql databases create moe_manager --instance=moe-manager-db
```

### 2. Secret Manager にシークレット登録

```bash
# DATABASE_URL（Cloud SQL の Unix socket 経由）
# 形式: postgres://USER:PASS@/moe_manager?host=/cloudsql/PROJECT:REGION:INSTANCE&sslmode=disable
echo -n "postgres://..." | gcloud secrets create DATABASE_URL --data-file=-
```

`TTS_SERVICE_URL` は TTS サービスをデプロイするまで作成しません。未設定時は API が `http://localhost:8001` にフォールバックし、音声機能のみ縮退します（`cmd/api/main.go`）。空文字の Secret はバージョンが作られず `:latest` 参照で起動に失敗するため、値が用意できるまで Secret 自体を作らず `deploy.yml` の `--set-secrets` からも外しておきます。

### 3. Workload Identity Federation 設定

GitHub Actions が長期キーなしで GCP 認証するための設定です。

```bash
# Workload Identity プール
gcloud iam workload-identity-pools create github-pool \
  --location=global --display-name="GitHub Actions Pool"

# OIDC プロバイダ（リポジトリを限定）
gcloud iam workload-identity-pools providers create-oidc github-provider \
  --location=global --workload-identity-pool=github-pool \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository=='mgmt-glasses/moe-manager'"

# デプロイ用サービスアカウント
gcloud iam service-accounts create github-deployer

# 必要なロールを付与（run.admin, artifactregistry.writer, cloudsql.client, iam.serviceAccountUser）
```

詳細は [google-github-actions/auth の README](https://github.com/google-github-actions/auth) を参照してください。

### 4. GitHub の Secrets / Variables 設定

**Secrets**（機密情報）

| 名前 | 値 |
|------|-----|
| `WIF_PROVIDER` | `projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github-pool/providers/github-provider` |
| `WIF_SERVICE_ACCOUNT` | `github-deployer@PROJECT_ID.iam.gserviceaccount.com` |

**Variables**（非機密設定）

| 名前 | 値の例 |
|------|--------|
| `GCP_PROJECT_ID` | `moe-manager` |
| `GCP_REGION` | `us-central1` |
| `CLOUD_SQL_INSTANCE` | `moe-manager:us-central1:moe-manager-db` |
| `VERTEX_LOCATION` | `us-central1` |
| `VERTEX_MODEL` | `gemini-2.5-flash` |

## デプロイ実行

デプロイは**手動実行のみ**です。課金を抑えるため push による自動デプロイは行いません。`deploy.yml` がデフォルトブランチ（`develop`）に存在する状態で有効になります（本 PR をマージ後に利用可能）。

### 手動デプロイ

GitHub の Actions タブから **Deploy to Cloud Run** ワークフローを `workflow_dispatch` で実行します。

本番（`main`）への自動デプロイは認証導入とあわせて別途整備します。

## 実行時環境変数

Cloud Run 側で設定される環境変数です（`deploy.yml` 参照）。

| 環境変数 | ソース | 説明 |
|----------|--------|------|
| `DATABASE_URL` | Secret Manager | Cloud SQL 接続文字列 |
| `TTS_SERVICE_URL` | Secret Manager | TTS サービスの URL（TTS デプロイ後に追加。未設定時は縮退） |
| `VERTEX_PROJECT` | Variables | GCP プロジェクト ID |
| `VERTEX_LOCATION` | Variables | Vertex AI のリージョン |
| `VERTEX_MODEL` | Variables | 使用する Gemini モデル |
| `PORT` | Cloud Run 自動注入 | リッスンポート（デフォルト 8080） |

## コスト管理

検証で使わない間は課金を抑えられます。

- **Cloud Run**: リクエストが無ければ自動でゼロインスタンスまで縮退し、ほぼ課金されません（`--min-instances` を 0 のままにする）。
- **Cloud SQL**: インスタンスは起動中ずっと課金されます。使わない間は停止します。

```bash
# Cloud SQL を停止（課金を抑える）
gcloud sql instances patch moe-manager-db --activation-policy=NEVER --project=moe-manager

# 再開（デプロイ前に起動しておく）
gcloud sql instances patch moe-manager-db --activation-policy=ALWAYS --project=moe-manager
```

停止中は API が DB に接続できず起動時マイグレーションも失敗するため、デプロイ前に必ず再開してください。

## 注意点

### 音声ファイルの永続化

現在の `VOICE_AUDIO_DIR`（デフォルト `data/voices`）はコンテナのローカルディスクに保存します。Cloud Run のディスクはエフェメラルなため、**インスタンス再起動・スケールで音声ファイルが消えます**。

- MVP では `--min-instances=1` かつ単一インスタンスで運用すれば当面は動作します
- 本番でスケールする場合は `AudioStorage` の GCS 実装を追加してください（interface は分離済み）

### マイグレーション

マイグレーションは `embed.FS` でバイナリに含まれ、起動時に `runMigrations` が自動適用します。別途実行は不要です。

### TTS サービス

TTS（PyTorch モデル）は GPU を要するため Cloud Run の標準インスタンスでは動きません。別途 GPU 対応環境にデプロイし、その URL を `TTS_SERVICE_URL` に設定してください。

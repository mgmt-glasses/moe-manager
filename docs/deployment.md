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

### 1. 必要な API の有効化

```bash
PROJECT_ID=moe-manager

gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com \
  secretmanager.googleapis.com \
  sqladmin.googleapis.com \
  iamcredentials.googleapis.com \
  sts.googleapis.com \
  aiplatform.googleapis.com \
  --project=$PROJECT_ID
```

`sqladmin`（Cloud SQL）と `cloudbuild`（`gcloud run deploy --source` のビルド）は忘れやすいので注意します。

### 2. GCP リソース作成

```bash
PROJECT_ID=moe-manager
REGION=us-central1

# Artifact Registry リポジトリ
gcloud artifacts repositories create moe-manager \
  --repository-format=docker --location=$REGION --project=$PROJECT_ID

# Cloud SQL (PostgreSQL) インスタンス
# デフォルト Edition は ENTERPRISE_PLUS で db-f1-micro を受け付けないため
# --edition=ENTERPRISE を明示する。
gcloud sql instances create moe-manager-db \
  --database-version=POSTGRES_16 --edition=ENTERPRISE \
  --tier=db-f1-micro --region=$REGION --project=$PROJECT_ID
gcloud sql databases create moe_manager --instance=moe-manager-db --project=$PROJECT_ID
gcloud sql users set-password postgres --instance=moe-manager-db \
  --password='<PASS>' --project=$PROJECT_ID
```

### 3. Secret Manager にシークレット登録

```bash
# DATABASE_URL（Cloud SQL の Unix socket 経由）
# 形式: postgres://USER:PASS@/moe_manager?host=/cloudsql/PROJECT:REGION:INSTANCE&sslmode=disable
echo -n "postgres://..." | gcloud secrets create DATABASE_URL --data-file=-

# GEMINI_API_KEY（スクリーンタイム画像解析で使用。未設定時は画像解析のみ縮退）
echo -n "AIza..." | gcloud secrets create GEMINI_API_KEY --data-file=-
```

`TTS_SERVICE_URL` は TTS サービスをデプロイするまで作成しません。未設定時は API が `http://localhost:8001` にフォールバックし、音声機能のみ縮退します（`cmd/api/main.go`）。空文字の Secret はバージョンが作られず `:latest` 参照で起動に失敗するため、値が用意できるまで Secret 自体を作らず `deploy.yml` の `--set-secrets` からも外しておきます。

### 4. IAM 設定（WIF・サービスアカウント・ロール）

GitHub Actions が長期キーなしで GCP 認証するための設定と、各サービスアカウントへのロール付与です。

```bash
PROJECT_ID=moe-manager
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
REPO=mgmt-glasses/moe-manager

# Workload Identity プール
gcloud iam workload-identity-pools create github-pool \
  --location=global --display-name="GitHub Actions Pool" --project=$PROJECT_ID

# OIDC プロバイダ（リポジトリを限定）
gcloud iam workload-identity-pools providers create-oidc github-provider \
  --location=global --workload-identity-pool=github-pool \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository=='${REPO}'" --project=$PROJECT_ID

# デプロイ用サービスアカウント
gcloud iam service-accounts create github-deployer --project=$PROJECT_ID
DEPLOYER_SA="github-deployer@${PROJECT_ID}.iam.gserviceaccount.com"

# デプロイ用 SA にロール付与
for ROLE in roles/run.admin roles/artifactregistry.writer roles/cloudsql.client \
            roles/iam.serviceAccountUser roles/secretmanager.secretAccessor; do
  gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${DEPLOYER_SA}" --role="$ROLE" --condition=None
done

# WIF からデプロイ用 SA を借用できるようにする
gcloud iam service-accounts add-iam-policy-binding "$DEPLOYER_SA" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/github-pool/attribute.repository/${REPO}"

# ランタイム SA（Cloud Run 実行時の既定 = compute SA）にロール付与
# Cloud Run が起動時に DB / Secret に到達し、Vertex AI を呼ぶために必要。
# cloudbuild.builds.builder は `gcloud run deploy --source` のビルドに使う。
COMPUTE_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"
for ROLE in roles/cloudsql.client roles/secretmanager.secretAccessor \
            roles/aiplatform.user roles/cloudbuild.builds.builder; do
  gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${COMPUTE_SA}" --role="$ROLE" --condition=None
done
```

詳細は [google-github-actions/auth の README](https://github.com/google-github-actions/auth) を参照してください。

### 5. GitHub の Secrets / Variables 設定

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

デプロイ前に Cloud SQL を起動しておきます（「コスト管理」参照）。

### 手動デプロイ（CI / 推奨）

GitHub の Actions タブから **Deploy to Cloud Run** ワークフローを `workflow_dispatch` で実行します。`deploy.yml` がデフォルトブランチ（`develop`）にある必要があります。

本番（`main`）への自動デプロイは認証導入とあわせて別途整備します。

### ローカルから直接デプロイ（CI を介さない検証用）

CI を待たずに手元から同じ構成でデプロイできます。`gcloud run deploy --source` は Cloud Build でビルドするため、ローカルに docker daemon は不要です。

```bash
gcloud run deploy moe-manager-api \
  --source . \
  --region us-central1 \
  --project moe-manager \
  --platform managed \
  --allow-unauthenticated \
  --add-cloudsql-instances moe-manager:us-central1:moe-manager-db \
  --set-env-vars VERTEX_PROJECT=moe-manager,VERTEX_LOCATION=us-central1,VERTEX_MODEL=gemini-2.5-flash \
  --set-secrets DATABASE_URL=DATABASE_URL:latest,GEMINI_API_KEY=GEMINI_API_KEY:latest
```

初回は Artifact Registry の `cloud-run-source-deploy` リポジトリが自動作成されます。`TTS_SERVICE_URL` は未デプロイのため付けません（付けると空 Secret 参照で失敗します）。

### 動作確認

デプロイ完了時に出力される Service URL に対してヘルスチェックします。ヘルスエンドポイントは `/api/v1/health` です。

```bash
URL=$(gcloud run services describe moe-manager-api --region us-central1 \
  --project moe-manager --format='value(status.url)')
curl -s "$URL/api/v1/health"      # => {"status":"ok"}
curl -s "$URL/api/v1/characters"  # => キャラクター一覧（DB 疎通確認）
```

`/api/v1/health` が 200、`/api/v1/characters` が実データを返せば、マイグレーション適用・DB 接続・API 稼働まで成功しています。

## デプロイ済みエンドポイント

- **Service URL**: `https://moe-manager-api-ydy2vlqfxa-uc.a.run.app`
- この URL は Cloud Run がプロジェクト・サービス単位で割り当てる固定値で、再デプロイ（リビジョン更新）しても変わりません。
- API は全て `/api/v1/` 配下にあります。例: `GET /api/v1/health`、`GET /api/v1/characters`。
- ルートパス `/` はハンドラを定義していないため `404 page not found` を返しますが、これは正常です（障害ではありません）。動作確認は必ず `/api/v1/health` など API パスで行ってください。
- 公開設定は `--allow-unauthenticated` のため、認証なしで誰でも到達できます（MVP 段階。認証は別途整備）。

## フロントエンド（iOS アプリ）からの接続手順

フロントエンドは iOS ネイティブアプリ（`moe-manger-frontend` の `MySecretary`）です。ブラウザを介さないため **CORS 設定は不要**です。

接続先はアプリ内の **開発者設定（DeveloperSettings）画面**で切り替えます（`APIEnvironment` が `UserDefaults` に永続化）。

- API モードは `mock` / `local` / `remote` の 3 種類で、初期値は `mock`。
- 実際の分岐は `mock` か否かの 2 通りです（`MySecretaryApp.swift`）。`mock` のときはリモートリポジトリを生成せず、モックデータで動作します。`mock` 以外（`local` / `remote`）のときは `baseURL` から `APIClient` を生成し、実 API に接続します。
- **`local` と `remote` に挙動差はありません**。どちらも接続先は `baseURL` で決まります（ラベル上の区別のみ）。
- baseURL の初期値は `http://localhost:8081/api/v1`（ローカル開発用）。

デプロイした Cloud Run の API に接続するには、開発者設定で次のように設定します。

1. iOS アプリを起動する。
2. 開発者設定（DeveloperSettings）画面を開く。
3. 接続モードを `mock` 以外に変更する。画面上の表示では **ローカル** または **リモート** のどちらでも可。
4. ベース URL に `https://moe-manager-api-ydy2vlqfxa-uc.a.run.app/api/v1` を設定する。末尾の `/api/v1` まで含める。
5. アプリを終了し、再起動する。接続設定は起動時に読み込まれるため、変更は次回起動時に反映される。

接続先は手順 4 のベース URL のみで決まります。アプリ側のコード変更は不要で、設定値の切り替えのみで完結します。

### フロントエンド接続前の疎通確認

iOS アプリから接続する前に、Cloud Run API が公開 URL で応答していることを確認します。

```bash
curl -s https://moe-manager-api-ydy2vlqfxa-uc.a.run.app/api/v1/health
curl -s https://moe-manager-api-ydy2vlqfxa-uc.a.run.app/api/v1/characters
```

期待結果:

- `/api/v1/health` が `{"status":"ok"}` を返す。
- `/api/v1/characters` が `success: true` とキャラクター一覧を返す。

上記が成功していれば、Cloud Run の起動、Cloud SQL 接続、起動時マイグレーション、公開 API の基本疎通は成功しています。iOS アプリは同じ HTTPS URL に直接アクセスするため、追加の CORS 設定やプロキシ設定は不要です。

## 実行時環境変数

Cloud Run 側で設定される環境変数です（`deploy.yml` 参照）。

| 環境変数 | ソース | 説明 |
|----------|--------|------|
| `DATABASE_URL` | Secret Manager | Cloud SQL 接続文字列 |
| `GEMINI_API_KEY` | Secret Manager | スクリーンタイム画像解析用の Gemini API キー（未設定時は画像解析のみ縮退） |
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

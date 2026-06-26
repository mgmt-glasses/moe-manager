# ADR-0010: ユーザー別 API の Firebase ID token 認証

## ステータス

Accepted

## 背景

MVP 初期はローカル開発を優先し、API は `userId` をパスで受け取るだけで認証を行っていなかった。
Cloud Run への公開（[ADR-0009](0009-deployment-cloud-run-cloud-sql.md)）に伴い、ユーザーのタスク・スクリーンタイム・チャット等の個人データが認証なしで誰でも取得できる状態になった。

フロントエンド（iOS）は Firebase Authentication でログインする方針のため、バックエンドはそのログインを信頼の起点にできる。
バックエンドに必要なのは「リクエストが正当な Firebase ユーザーから来ていること」と「他人の `userId` のデータにアクセスしていないこと」の保証である。

## 判断

ユーザー別 API に Firebase ID token による認証を必須とする。

- 認証は `internal/auth` のミドルウェアで行う横断的関心事とし、各ドメインには持ち込まない。
- ID token は Firebase Admin SDK を使わず、`golang-jwt` で RS256 署名を自前検証する。Google の公開証明書（`securetoken@system` の x509）を取得し、`Cache-Control: max-age` に従ってキャッシュする。
- 検証では署名に加えて issuer（`https://securetoken.google.com/<projectID>`）、audience（projectID）、subject（`uid`）を確認する。
- 認可は二段階で行う。
  - `RequireAuth`: `Authorization: Bearer <ID token>` を検証し、検証済みユーザーを context に載せる。失敗時は `401 UNAUTHORIZED`。
  - `RequirePathUser`: パスの `{userId}` と token の `uid` を照合する。不一致時は `403 FORBIDDEN`。
- 公開ルートは `GET /api/v1/health` と `GET /api/v1/characters`（一覧・詳細）のみ。それ以外の `/api/v1/users/...` 配下はすべて認証必須とする。
- 検証対象プロジェクトは環境変数 `FIREBASE_PROJECT_ID` で指定する。

## 理由

- **Firebase ID token を採用**: フロントが Firebase Auth を使うため、ログイン基盤を再実装せず信頼を引き継げる。
- **Admin SDK ではなく自前検証**: 必要なのは ID token 検証のみで、Admin SDK が要求するサービスアカウント鍵や追加依存を避けたい。標準 JWT ライブラリと公開証明書だけで完結し、横断層を軽量に保てる。証明書はキャッシュするため検証ごとの外部呼び出しは発生しない。
- **ミドルウェア層に配置**: 認証・認可は全ドメイン共通の関心事であり、`model`/`service` を HTTP やトークンの知識から独立させる依存ルール（[architecture-guidelines](../architecture-guidelines.md)）に従う。検証済みユーザーは context で渡し、handler は primitive な `uid` だけを受け取る。
- **401 と 403 の分離**: 「未認証（トークンが無い・不正）」と「認証済みだが他人のリソース」を区別し、クライアントが再ログインと権限エラーを正しく扱えるようにする。
- **却下した代替案**:
  - Cloud Run IAM のみで保護 → サービス単位の認可しかできず、ユーザー単位の `userId` 所有権チェックができない。
  - API キー方式 → ユーザーを識別できず、フロントの既存ログインと二重管理になる。

## 結果

- すべてのユーザー別 API 呼び出しに有効な Firebase ID token が必要になる。クライアントはログイン後に token を付与する。
- `FIREBASE_PROJECT_ID` が未設定だとサーバー起動に失敗する（フェイルクローズ）。ローカル・Cloud Run 双方で設定が必須。
- ローカル開発・テスト用の例外として、`AUTH_BYPASS=true` を設定した場合に限り、署名検証を行わず Bearer トークン文字列をそのまま `uid` として扱う `BypassVerifier` を `FirebaseVerifier` の代わりに注入する。`RequireAuth` / `RequirePathUser` のロジックは変更せず、`TokenVerifier` の差し替えだけで実現する。フロントの Firebase ログインが未実装の間も認証必須 API を試せるようにするための措置であり、本番では `AUTH_BYPASS` を設定しない（設定すると任意の `uid` でのなりすましを許す）。
- 認証仕様の正は `internal/auth` の実装と [api-reference](../api-reference.md) / [api-guidelines](../api-guidelines.md) とする。
- トークン失効（revoke）の即時反映やカスタムクレームによるロール認可は本 ADR の範囲外とし、必要になった時点で別途検討する。
- 公開証明書キャッシュの可用性は次の方針で硬化する。期限切れ時の再取得は singleflight で 1 本に集約する（thundering herd 回避）。取得失敗時は既存 cert を短い猶予の間 stale のまま使い続け、再取得を間引く（cert エンドポイント一時障害での全ログイン不能を回避）。直前世代の cert を 1 世代保持し、ローテーション重複期間中の旧 kid トークンも検証できるようにする。

# テスト方針

このドキュメントは、Go バックエンドにおけるテストコードの目的、背景、実装方針をまとめます。
2026-06 に `character` / `user` / `task` / `statistics` / `chat` の各ドメインへテストを追加した際に定めた方針です。

## 目的

- API レスポンス（ステータスコード、JSON のキー名・値）が `docs/Issue/*.md` の仕様通りであることをテストで保証する。
- domain 層のビジネスロジック（バリデーション、認可、エラー伝播）を外部依存なしで検証できる状態にする。
- 新規ドメインを追加する際に、同じ構成でテストを書けるようにする。

## 背景

テスト追加前は以下の状態だった。

- `character` / `user` / `task` / `statistics` は `go test ./...` で `no test files` となり、テストが一切なかった。
- `chat` のみ `service_test.go` があったが、HTTP handler のテストはなかった。
- `cmd/api/main.go` を確認すると `chat` ハンドラがルーティングに登録されておらず未配線だった。

これらを解消するため、ドメインごとに 1 PR で fake・service テスト・handler テストを追加した。

## 対象範囲

### やったこと

- `internal/<domain>/testutil/` に Repository / Query などの interface を実装する in-memory fake を追加する。
- `internal/<domain>/service_test.go` に business logic の単体テストを追加する。
- `internal/<domain>/handler_test.go` に HTTP handler のテストを追加する。

### やらないこと

- `internal/<domain>/adapter/` の PostgreSQL 実装に対する統合テスト（実 DB 接続）は対象外とする。
- 未実装のドメイン（`screentime`、`voice`、`chat-logs`）のテストは対象外とする。
- `chat` の `cmd/api/main.go` 配線は対象外とする。理由は [既知の制限](#既知の制限) を参照。

## 実装方針

### ファイル配置

```txt
internal/<domain>/
├── testutil/
│   └── fake_<interface>.go   # Repository / Query などの in-memory fake
├── service_test.go           # service 層の単体テスト
└── handler_test.go           # HTTP handler のテスト
```

`internal/chat/testutil/` は本 Issue 以前から存在していたため、同じ命名・実装パターンを他ドメインにも適用した。

### fake の実装ルール

- fake は対象ドメインが定義する interface（`Repository`、`CharacterValidator`、`StatisticsQuery` など）を実装する。
- フィールドに `XxxErr error` を持たせ、任意のメソッドでエラーを注入できるようにする。
- 永続化が必要なもの（`Repository` 系）は `map[string]T` で in-memory に保持する。
- 本物の DB アダプタの挙動に合わせる（例: 存在しない ID は `ErrNotFound` を返す）。

### service 層テスト

- パッケージは `<domain>_test`（外部テストパッケージ）にし、公開 API のみを経由してテストする。
- 正常系、入力バリデーションエラー、認可エラー（他ユーザーのリソース操作など）、依存先（repository/query）のエラー伝播を確認する。

### handler 層テスト

- `net/http/httptest` と実際の `go-chi/chi` ルーターを使い、`cmd/api/main.go` と同じルーティング構造（`r.Route("/api/v1/users/{userId}", ...)` など）を各テストファイル内に再現する。
- 実際に HTTP リクエストを発行し、レスポンスのステータスコードと JSON ボディを検証する。
- 共通レスポンス形式 `{"success", "data", "error"}`（`docs/api-guidelines.md` 参照）をデコードする `envelope` 型をテストファイルごとに定義し、`data` の中身は domain ごとの期待値と比較する。
- 確認する観点:
  - 正常系のレスポンス構造とフィールド名が仕様（`docs/Issue/*.md`）と一致するか。
  - バリデーションエラー・Not Found・Forbidden などの異常系で、仕様通りのステータスコードとエラーコードが返るか。
  - 空配列や未登録データなど、レスポンス形式が安定しているか。

### 使用するツール

- 標準ライブラリの `testing` パッケージのみを使う。外部アサーションライブラリ（testify 等）は導入しない。既存の `internal/chat` のテスト規約に合わせている。

## 実装内容

| ドメイン | 追加ファイル | 主な確認内容 | PR |
| --- | --- | --- | --- |
| `character` | `testutil/fake_repository.go`, `service_test.go`, `handler_test.go` | キャラ一覧/詳細のレスポンス形、`voiceId`/`iconUrl` 等のフィールド変換、404/500 | [#24](https://github.com/mgmt-glasses/moe-manager/pull/24) |
| `user` | `testutil/fake_repository.go`, `testutil/fake_character_validator.go`, `service_test.go`, `handler_test.go` | Create/Update/選択キャラ更新のバリデーション、存在しないキャラ/ユーザー、201/200/400/404 | [#25](https://github.com/mgmt-glasses/moe-manager/pull/25) |
| `task` | `testutil/fake_repository.go`, `service_test.go`, `handler_test.go` | CRUD、他ユーザーのタスク操作禁止（403）、201/200/400/404 | [#26](https://github.com/mgmt-glasses/moe-manager/pull/26) |
| `statistics` | `testutil/fake_query.go`, `service_test.go`, `handler_test.go` | today/daily/weekly のネスト/フラット構造、completionRate・diffMinutes 計算、日付バリデーション、400/500 | [#27](https://github.com/mgmt-glasses/moe-manager/pull/27) |
| `chat` | `handler_test.go`（既存 `testutil`、`service_test.go` を再利用） | チャット送信のレスポンス形、キャラ未選択時 422、LLM/コンテキスト取得エラー時 500 | [#28](https://github.com/mgmt-glasses/moe-manager/pull/28) |

## 既知の制限

`internal/chat` には `LLMClient` / `ContextLoader` / `PersonaRepository` / `ChatLogger` の本番 adapter（`internal/chat/adapter/`）がまだ実装されていない。
そのため `cmd/api/main.go` への `chat` ハンドラの配線は本対応では行わず、別 Issue（LLM 連携・他ドメイン横断のコンテキスト取得の実装）に持ち越している。

## 今後の対応

- `screentime`、`voice`、`chat-logs` の各ドメインを実装する際は、本ドキュメントと同じ構成（`testutil` fake、`service_test.go`、`handler_test.go`）でテストを追加する。
- `internal/chat/adapter` の実装後、`cmd/api/main.go` への配線とそれに伴う統合確認を別 Issue で行う。

## 参照ドキュメント

- `docs/api-guidelines.md`
- `docs/implementation-guide.md`
- `docs/Issue/01-character-selection.md`
- `docs/Issue/02-character-chat.md`
- `docs/Issue/04-task-management.md`
- `docs/Issue/06-statistics-summary.md`

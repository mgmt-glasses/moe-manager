# タスクを登録・完了できる

## 目的

ユーザがタスクを登録し、一覧確認、完了、再オープン、削除を行えるようにする。
タスク達成状況は統計とチャットコンテキストで利用できる状態にする。

## 背景

MVP では、秘書キャラがユーザの行動に反応するための主要データとしてタスク情報を扱う。
タスク管理は `internal/task` に閉じ、他ドメインからは利用側が定義する Query interface 経由で参照する。

## 対象範囲

### やること

- タスクを登録できる。
- ユーザごとのタスク一覧を取得できる。
- タスクを完了、再オープンできる。
- タスクを削除できる。
- 完了率計算に必要な状態を保持する。

### やらないこと

- 複雑なステータス管理は扱わない。
- タスクのタグ、優先度、期限通知は扱わない。
- 統計表示そのものは別 Issue で扱う。

## 実装方針

- 対象モジュール: `task`, `api`
- 想定ブランチ: `feature/task-management`
- 主な変更ファイル/ディレクトリ:
  - `internal/task/`
  - `cmd/api/`
  - `prisma/migrations/`
- 依存する Issue: なし
- 後続 Issue: `02-character-chat.md`, `06-statistics-summary.md`

実装は domain model、interface、use case、repository adapter、HTTP handler、`cmd/api` での登録の順で進める。

## 受け入れ条件

- [ ] ユーザがタスクを登録できる。
- [ ] ユーザごとのタスク一覧を取得できる。
- [ ] 未完了タスクを完了にできる。
- [ ] 完了タスクを未完了に戻せる。
- [ ] タスクを削除できる。
- [ ] 他ユーザのタスクを操作できない。

## API / DB 変更

### API

- `POST /api/v1/users/{userId}/tasks`
- `GET /api/v1/users/{userId}/tasks`
- `PATCH /api/v1/users/{userId}/tasks/{taskId}/complete`
- `PATCH /api/v1/users/{userId}/tasks/{taskId}/reopen`
- `DELETE /api/v1/users/{userId}/tasks/{taskId}`

### DB

- `tasks` テーブルを扱う。
- MVP のタスク状態は `todo` / `done` を基本とする。
- 作成日時、完了日時、削除扱いの方針を実装前に確認する。

## テスト・動作確認

- [ ] タスク登録の正常系とバリデーションエラーを確認する。
- [ ] 一覧取得でユーザごとのタスクのみ返ることを確認する。
- [ ] 完了、再オープン、削除の正常系を確認する。
- [ ] 存在しないタスク ID の異常系を確認する。

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/api-guidelines.md`
- `docs/database-guidelines.md`
- `docs/implementation-guide.md`
- `docs/detail/db-design.md`

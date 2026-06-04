# タスク達成状況と娯楽時間を統計として確認できる

## 目的

ユーザが今日と過去 7 日分のタスク達成状況、娯楽時間、日次サマリーを確認できるようにする。
統計は専用集計テーブルを持たず、既存データから都度計算する。

## 背景

MVP では、タスクと娯楽時間の状況を可視化し、チャットでの秘書キャラの反応にも利用できるようにする。
`internal/statistics` は例外的に他ドメインが管理するテーブルを Query adapter から横断して集計できる。

## 対象範囲

### やること

- 今日のタスク完了数、未完了数、完了率を取得できる。
- 今日の娯楽時間と目標との差分を取得できる。
- 日次サマリーを取得できる。
- 過去 7 日分のサマリーを取得できる。

### やらないこと

- 専用の統計保存テーブルは作らない。
- 高度なグラフ分析や月次集計は扱わない。
- タスクや娯楽時間の CRUD 自体は別 Issue で扱う。

## 実装方針

- 対象モジュール: `statistics`, `api`
- 想定ブランチ: `feature/statistics-summary`
- 主な変更ファイル/ディレクトリ:
  - `internal/statistics/`
  - `cmd/api/`
- 依存する Issue: `04-task-management.md`, `05-screentime-recording.md`
- 後続 Issue: `02-character-chat.md`

集計は Query interface と adapter で扱い、service が他モジュールの内部実装に依存しないようにする。

## 受け入れ条件

- [ ] 今日の統計を取得できる。
- [ ] 指定日の統計を取得できる。
- [ ] 過去 7 日分の週次統計を取得できる。
- [ ] タスク未登録や娯楽時間未登録の日でもレスポンス形式が安定している。
- [ ] 統計専用テーブルを追加せず、既存記録から計算している。

## API / DB 変更

### API

- `GET /api/v1/users/{userId}/stats/today`
- `GET /api/v1/users/{userId}/stats/daily/{date}`
- `GET /api/v1/users/{userId}/stats/weekly`

### DB

- 新規の統計保存テーブルは追加しない。
- `tasks` と `screentime_records` を参照して都度計算する。

## テスト・動作確認

- [ ] 今日の統計取得の正常系を確認する。
- [ ] 指定日統計取得の正常系を確認する。
- [ ] 過去 7 日分統計取得の正常系を確認する。
- [ ] データが存在しない日のレスポンスを確認する。
- [ ] 完了率と娯楽時間差分の計算を確認する。

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/api-guidelines.md`
- `docs/database-guidelines.md`
- `docs/architecture-guidelines.md`
- `docs/detail/api_design_mvp.md`

# スクリーンタイム、または娯楽時間を記録できる

## 目的

ユーザが日付ごとの娯楽時間を手動で記録し、目標時間との差分を確認できるようにする。
記録データは統計とチャットコンテキストで利用できる状態にする。

## 背景

MVP ではスクリーンタイムの自動取得ではなく、娯楽時間の手動入力から始める。
DB テーブル名は `screentime_records` を正とし、旧表記の `entertainment_records` / `entertainment_times` は使わない。

## 対象範囲

### やること

- 日付ごとの娯楽時間を登録・更新できる。
- 特定日の娯楽時間を取得できる。
- 期間指定で娯楽時間を取得できる。
- ユーザの目標娯楽時間との差分を算出できる。

### やらないこと

- OS のスクリーンタイム自動取得は扱わない。
- アプリ別利用時間の詳細分類は扱わない。
- 統計画面の集計表示は別 Issue で扱う。

## 実装方針

- 対象モジュール: `screentime`, `api`
- 想定ブランチ: `feature/screentime-recording`
- 主な変更ファイル/ディレクトリ:
  - `internal/screentime/`
  - `cmd/api/`
  - `migrations/`
- 依存する Issue: ユーザ目標時間を扱う場合はユーザ設定 API
- 後続 Issue: `02-character-chat.md`, `06-statistics-summary.md`

API パスは既存方針に合わせて `entertainment-records` を使う可能性があるが、DB/domain の正規名は `screentime_records` とする。

## 受け入れ条件

- [ ] ユーザが日付ごとの娯楽時間を保存できる。
- [ ] 同じ日付の記録は更新として扱える。
- [ ] 特定日の記録を取得できる。
- [ ] 期間指定で記録一覧を取得できる。
- [ ] 目標時間との差分を確認できる。
- [ ] 日付形式は `YYYY-MM-DD` として扱う。

## API / DB 変更

### API

- `PUT /api/v1/users/{userId}/entertainment-records/{date}`
- `GET /api/v1/users/{userId}/entertainment-records/{date}`
- `GET /api/v1/users/{userId}/entertainment-records`

### DB

- `screentime_records` テーブルを扱う。
- ユーザ ID、対象日、娯楽時間、作成日時、更新日時を保存する。
- 同一ユーザ・同一日付の重複登録を避ける。

## テスト・動作確認

- [ ] 記録保存の正常系を確認する。
- [ ] 同一日の更新が重複登録にならないことを確認する。
- [ ] 特定日取得、期間取得の正常系を確認する。
- [ ] 日付形式や時間値のバリデーションを確認する。

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/api-guidelines.md`
- `docs/database-guidelines.md`
- `docs/documentation-guidelines.md`
- `docs/detail/db-design.md`

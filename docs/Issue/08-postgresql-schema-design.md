# MVP PostgreSQL スキーマを設計できる

## 目的

MVP で必要なデータを PostgreSQL に保存するため、実装者ごとに解釈が分かれない粒度で DB スキーマを確定する。
この Issue では migration ツールや ORM を新たに選定せず、後続の SQL migration 実装が参照する設計を作る。

## 背景

MVP では、キャラクター選択、ユーザ設定、タスク、娯楽時間、チャットログ、音声ファイルを永続化する。
既存の `docs/detail/db-design.md` には主要テーブルとカラムが定義されているが、NULL 可否、デフォルト値、外部キーの削除時挙動、制約など、SQL を実装するために必要な判断が一部未確定である。

各機能ブランチが個別判断で migration を追加すると、同じテーブルに異なる制約や命名が導入される可能性がある。
そのため、MVP の PostgreSQL スキーマ設計を先に確定し、後続実装の正本とする。

## 対象範囲

### やること

- MVP で必要なテーブルと関連を確定する。
- 各カラムの PostgreSQL 型、NULL 可否、デフォルト値を確定する。
- 主キー、外部キー、unique 制約、check 制約を確定する。
- 外部キー参照先が削除された場合の挙動を確定する。
- MVP の主要クエリに必要なインデックスを確定する。
- ID と日時の生成・更新方針を確定する。
- 初期キャラクターデータの管理方針を確定する。
- 設計結果を `docs/detail/db-design.md` と `docs/database-guidelines.md` に反映する。

### やらないこと

- SQL migration ファイルの作成・適用は扱わない。
- migration ライブラリや ORM の新規選定・導入は扱わない。
- Go の repository adapter や API は実装しない。
- 本番 DB、バックアップ、監視、認証・認可は設計しない。
- 統計専用テーブルは作成しない。
- 長期記憶、会話要約、好感度、関係性 state、ユーザプロファイル抽出用のテーブルは設計しない。

## 実装方針

- 対象モジュール: `db`, `docs`
- 想定ブランチ: `feature/db-postgres-schema-design`
- 主な変更ファイル:
  - `docs/detail/db-design.md`
  - `docs/database-guidelines.md`
  - `docs/Issue/README.md`
- 依存する Issue: なし
- 後続 Issue: PostgreSQL migration 実装、各機能の repository adapter 実装

PostgreSQL の機能と型を前提に設計する。
設計上の正本は `docs/detail/db-design.md` とし、`docs/database-guidelines.md` にはチーム全体で守る運用方針を記載する。
後続の migration 実装では、確定した設計を SQL の `up` / `down` migration に反映する。

## 設計対象

### テーブル

- `characters`
- `users`
- `tasks`
- `screentime_records`
- `chat_logs`
- `voice_files`

### 確定が必要な項目

- 文字列 ID の生成主体と形式
- `created_at`、`updated_at` のデフォルト値と更新主体
- `users.selected_character_id` の NULL 可否と削除時挙動
- `tasks.status` と `chat_logs.role` の許容値
- タスク完了時・未完了時の `completed_at` の整合性
- 娯楽時間と目標時間に許可する値の範囲
- `screentime_records(user_id, date)` の重複可否
- `chat_logs.voice_file_id` の NULL 可否と削除時挙動
- ユーザ、キャラクター、音声ファイル削除時のログ保持方針
- 初期キャラクターデータを seed として管理する方法

## 受け入れ条件

- [ ] MVP の全テーブルについてカラム、型、NULL 可否、デフォルト値が定義されている。
- [ ] 全ての主キー、外部キー、unique 制約、check 制約が定義されている。
- [ ] 全ての外部キーについて更新・削除時の挙動が定義されている。
- [ ] MVP の主要クエリと、それを支えるインデックスの対応が説明されている。
- [ ] ID、日時、初期データの管理方針が定義されている。
- [ ] 統計専用テーブルを持たない方針と矛盾していない。
- [ ] 後続担当者が追加判断なしで SQL migration を実装できる。
- [ ] Prisma や Node.js を前提とする記述が含まれていない。

## API / DB 変更

### API

- API の変更は行わない。

### DB

- この Issue では DB に変更を適用しない。
- 後続 Issue で SQL migration として適用するスキーマを確定する。

## テスト・動作確認

- [ ] 各 MVP 機能の保存・取得要件が設計したテーブルで満たせることを確認する。
- [ ] 外部キーと削除時挙動に循環や矛盾がないことを確認する。
- [ ] unique 制約と check 制約が MVP の操作を妨げないことを確認する。
- [ ] 主要クエリが定義したインデックスを利用できる構造になっていることを確認する。
- [ ] 既存の各機能 Issue と DB 設計の命名が一致していることを確認する。

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/database-guidelines.md`
- `docs/detail/db-design.md`
- `docs/Issue/01-character-selection.md`
- `docs/Issue/04-task-management.md`
- `docs/Issue/05-screentime-recording.md`
- `docs/Issue/07-chat-logs.md`

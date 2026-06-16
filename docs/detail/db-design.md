# 二次元秘書AIタスク・娯楽管理アプリ DB設計

## 1. 目的

このドキュメントは、二次元秘書AIタスク・娯楽管理アプリのMVPで使用するDB設計を定義する。

DBは、ユーザ設定の保存、タスク管理、娯楽時間の追跡、および秘書キャラとのチャット履歴の永続化を担う。「社長（ユーザ）の行動を秘書が把握し、適切に反応できる」状態を維持するためのデータ基盤とする。

この設計は、後続の SQL migration 実装および repository adapter 実装が参照する正本であり、実装者ごとの追加判断を必要としない粒度で確定する。

## 2. 基本方針

MVPでは、本番環境との一貫性を重視し、PostgreSQLを採用する。ローカル開発ではDockerでPostgreSQLを起動する。

方針:
- **マイグレーション管理**: スキーマ変更は `migrations/` の SQL migration ファイル（`golang-migrate`）で管理する。
- **マスターとトランザクションの分離**: キャラクター定義などの「マスターデータ」と、タスクやログなどの「ユーザデータ」を明確に分ける。
- **統計の即時性**: MVPでは専用の統計テーブルは持たず、タスクや娯楽時間の記録から都度計算する方針とする。
- **AIコンテキストの最適化**: チャットログは、AIが参照しやすい粒度で保存し、最新の履歴を迅速に取得できるようにする。

## 3. データ分類

DB内のデータは以下のカテゴリに分類する。

| 分類 | 例 | 方針 |
| --- | --- | --- |
| **ユーザ設定** | ユーザ名、社長としての呼び名、目標娯楽時間 | アプリの基本動作を決定するデータ |
| **キャラクター定義** | MBTI別キャラ名、性格、口調、ボイス設定 | アプリに同梱される静的なマスターデータ |
| **タスク管理** | タスク内容、完了フラグ、作成/完了日時 | ユーザの主活動を記録するデータ |
| **活動記録** | 日別の娯楽時間（スクリーンタイム） | 統計の元となるトランザクションデータ |
| **チャットログ** | 会話内容、発言者、日時 | 過去の文脈を維持するための履歴データ |
| **音声ファイル** | 生成テキスト、保存パス | TTS結果の参照に使う履歴データ |

## 4. ID・日時の生成方針

- **マスターデータ（`characters`）**: IDは固定文字列とし、アプリ同梱データとして migration の `INSERT` で投入する（例: `char_istj_001`）。アプリケーションコードから新規作成・削除しない。
- **ユーザデータ（`users`, `tasks`, `screentime_records`, `chat_logs`, `voice_files`）**: IDはアプリケーション層で UUID v4 を生成し（`uuid.NewString()`）、`TEXT` 型で保存する。DB側のデフォルト生成（`gen_random_uuid()` 等）は使わない。
- **`created_at`**: 全テーブルで `TIMESTAMPTZ NOT NULL DEFAULT NOW()`。行作成時に確定し、以降変更しない。
- **`updated_at`**: `characters`, `users`, `screentime_records` のみが持つ。`TIMESTAMPTZ NOT NULL DEFAULT NOW()` とし、UPDATE 時は repository 層が明示的に `NOW()` をSETする。DBトリガー（`ON UPDATE` 相当）は使わない。
- **`tasks`, `chat_logs`, `voice_files`**: イミュータブルなレコードとして扱い、`updated_at` を持たない。

## 5. マスターテーブル

### characters

秘書キャラクターの定義を保持する。アプリケーションから削除されない静的マスターデータ。

| カラム | 型 | NULL | デフォルト | 説明 |
| --- | --- | --- | --- | --- |
| id | TEXT | NOT NULL | - | キャラクターID（例: `char_istj_001`） |
| name | TEXT | NOT NULL | - | キャラクター名 |
| mbti_type | TEXT | NOT NULL | - | MBTIタイプ（例: `ISTJ`, `ENFJ`） |
| personality_desc | TEXT | NOT NULL | - | 性格説明（AIプロンプト用） |
| speech_style | TEXT | NOT NULL | - | 口調の定義（AIプロンプト用） |
| system_prompt_fragment | TEXT | NOT NULL | - | AI用のシステムプロンプト断片 |
| voice_preset_id | TEXT | NOT NULL | - | 使用するボイスプリセットのID |
| icon_path | TEXT | NOT NULL | `''` | アイコン画像のパス |
| standing_image_path | TEXT | NOT NULL | `''` | 立ち絵画像のパス |
| sample_voice_path | TEXT | NOT NULL | `''` | サンプルボイスのパス |
| created_at | TIMESTAMPTZ | NOT NULL | `NOW()` | 作成日時 |
| updated_at | TIMESTAMPTZ | NOT NULL | `NOW()` | 更新日時 |

制約:

- 主キー: `id`
- 外部キー、unique、check制約: なし

## 6. ユーザ・設定テーブル

### users

ユーザの基本プロフィールと現在設定を保持する。

| カラム | 型 | NULL | デフォルト | 説明 |
| --- | --- | --- | --- | --- |
| id | TEXT | NOT NULL | - | ユーザID（UUID v4） |
| name | TEXT | NOT NULL | - | 本名/ユーザ名 |
| president_name | TEXT | NOT NULL | - | キャラクターからの呼び名（例: `社長`） |
| selected_character_id | TEXT | NULL可 | `NULL` | 現在選択中のキャラクターID |
| target_entertainment_minutes | INTEGER | NOT NULL | `120` | 1日の目標娯楽時間（分） |
| created_at | TIMESTAMPTZ | NOT NULL | `NOW()` | 作成日時 |
| updated_at | TIMESTAMPTZ | NOT NULL | `NOW()` | 更新日時 |

制約:

- 主キー: `id`
- 外部キー: `selected_character_id` → `characters.id`（`ON DELETE SET NULL`, `ON UPDATE CASCADE`）
  - キャラクター未選択の状態を許容するため `selected_character_id` は `NULL` を許可する。
  - MVP では `characters` の削除APIを設けないため `SET NULL` が発火することは想定しないが、将来のキャラ廃止に備えて定義する。
- check制約: `target_entertainment_minutes BETWEEN 0 AND 1440`

## 7. タスク・活動テーブル

### tasks

ユーザが登録したタスクを管理する。

| カラム | 型 | NULL | デフォルト | 説明 |
| --- | --- | --- | --- | --- |
| id | TEXT | NOT NULL | - | タスクID（UUID v4） |
| user_id | TEXT | NOT NULL | - | ユーザID |
| title | TEXT | NOT NULL | - | タスク内容 |
| status | TEXT | NOT NULL | `'todo'` | 状態（`todo`, `done`） |
| created_at | TIMESTAMPTZ | NOT NULL | `NOW()` | 登録日時 |
| completed_at | TIMESTAMPTZ | NULL可 | `NULL` | 完了日時 |
| deleted_at | TIMESTAMPTZ | NULL可 | `NULL` | 論理削除日時 |

制約:

- 主キー: `id`
- 外部キー: `user_id` → `users.id`（`ON DELETE CASCADE`, `ON UPDATE CASCADE`）
  - ユーザ削除時はタスクも削除する。
- check制約: `status IN ('todo', 'done')`
- check制約: `(status = 'todo' AND completed_at IS NULL) OR (status = 'done' AND completed_at IS NOT NULL)`
  - 未完了タスクに完了日時が入っている状態、完了タスクに完了日時が無い状態を防ぐ。
- 削除はアプリケーション層で `deleted_at` を設定する論理削除とする（既存実装どおり）。物理削除は行わない。

### screentime_records

手動入力された娯楽時間を記録する。

| カラム | 型 | NULL | デフォルト | 説明 |
| --- | --- | --- | --- | --- |
| id | TEXT | NOT NULL | - | 記録ID（UUID v4） |
| user_id | TEXT | NOT NULL | - | ユーザID |
| date | DATE | NOT NULL | - | 対象日（`YYYY-MM-DD`） |
| minutes | INTEGER | NOT NULL | - | 娯楽時間（分） |
| target_minutes_snapshot | INTEGER | NOT NULL | - | 記録時点の目標時間（分、統計用スナップショット） |
| created_at | TIMESTAMPTZ | NOT NULL | `NOW()` | 記録日時 |
| updated_at | TIMESTAMPTZ | NOT NULL | `NOW()` | 更新日時 |

制約:

- 主キー: `id`
- 外部キー: `user_id` → `users.id`（`ON DELETE CASCADE`, `ON UPDATE CASCADE`）
- unique制約: `(user_id, date)` — 同一ユーザ・同一日付の重複記録を禁止する。`PUT /entertainment-records/{date}` は `ON CONFLICT (user_id, date) DO UPDATE` で upsert する。
- check制約: `minutes BETWEEN 0 AND 1440`
- check制約: `target_minutes_snapshot BETWEEN 0 AND 1440`

## 8. ログテーブル

migration 実装順序として、`voice_files` は `chat_logs` より先に作成する（`chat_logs.voice_file_id` が `voice_files.id` を参照するため）。

### voice_files

生成された音声ファイルのメタデータを管理する。

| カラム | 型 | NULL | デフォルト | 説明 |
| --- | --- | --- | --- | --- |
| id | TEXT | NOT NULL | - | ファイルID（UUID v4） |
| character_id | TEXT | NOT NULL | - | 生成に使用したキャラID |
| text | TEXT | NOT NULL | - | 読み上げ元テキスト |
| file_path | TEXT | NOT NULL | - | ローカルの保存パス |
| created_at | TIMESTAMPTZ | NOT NULL | `NOW()` | 生成日時 |

制約:

- 主キー: `id`
- 外部キー: `character_id` → `characters.id`（`ON DELETE RESTRICT`, `ON UPDATE CASCADE`）
  - MVP では `characters` の削除APIを設けないため、`RESTRICT` により誤削除を検知できる状態にする。

### chat_logs

ユーザとAIの会話履歴を保持する。

| カラム | 型 | NULL | デフォルト | 説明 |
| --- | --- | --- | --- | --- |
| id | TEXT | NOT NULL | - | ログID（UUID v4） |
| user_id | TEXT | NOT NULL | - | ユーザID |
| character_id | TEXT | NOT NULL | - | 会話相手のキャラID |
| role | TEXT | NOT NULL | - | 発言者（`user`, `assistant`） |
| message | TEXT | NOT NULL | - | 発言内容 |
| voice_file_id | TEXT | NULL可 | `NULL` | 生成されたボイスファイルのID |
| created_at | TIMESTAMPTZ | NOT NULL | `NOW()` | 発言日時 |

制約:

- 主キー: `id`
- 外部キー: `user_id` → `users.id`（`ON DELETE CASCADE`, `ON UPDATE CASCADE`）
  - ユーザ削除時はチャットログも削除する。
- 外部キー: `character_id` → `characters.id`（`ON DELETE RESTRICT`, `ON UPDATE CASCADE`）
- 外部キー: `voice_file_id` → `voice_files.id`（`ON DELETE SET NULL`, `ON UPDATE CASCADE`）
  - 音声ファイルが削除されてもログ本文は保持し、音声参照のみ失う。
- check制約: `role IN ('user', 'assistant')`
- check制約: `voice_file_id IS NULL OR role = 'assistant'`
  - ボイスはキャラ返答（`assistant`）にのみ紐づく。

保存タイミング:

- ユーザ発言は LLM 呼び出し前に `voice_file_id = NULL` で保存する。
- キャラ返答は生成成功後に `voice_file_id = NULL` で保存する。
- ボイス生成に成功した場合、保存済みのキャラ返答行に対して `voice_file_id` を `UPDATE` で関連付ける（`AttachVoice`）。

日付による絞り込みは JST の暦日として扱い、DB には `timestamptz`（UTC）で保存する。日付変換はアプリケーション層で行う。

## 9. 外部キー一覧と削除時挙動

| テーブル.カラム | 参照先 | ON DELETE | ON UPDATE | NULL可否 |
| --- | --- | --- | --- | --- |
| `users.selected_character_id` | `characters.id` | SET NULL | CASCADE | NULL可 |
| `tasks.user_id` | `users.id` | CASCADE | CASCADE | NOT NULL |
| `screentime_records.user_id` | `users.id` | CASCADE | CASCADE | NOT NULL |
| `chat_logs.user_id` | `users.id` | CASCADE | CASCADE | NOT NULL |
| `chat_logs.character_id` | `characters.id` | RESTRICT | CASCADE | NOT NULL |
| `chat_logs.voice_file_id` | `voice_files.id` | SET NULL | CASCADE | NULL可 |
| `voice_files.character_id` | `characters.id` | RESTRICT | CASCADE | NOT NULL |

`characters` を参照する外部キーに循環は無く、`characters` の削除は MVP では発生しない（削除APIなし）ため `RESTRICT` が運用上の問題になることはない。
`users` の削除は関連する `tasks`, `screentime_records`, `chat_logs` を `CASCADE` で連鎖削除する。

## 10. インデックス方針

主要クエリと対応するインデックス:

| インデックス | 用途 |
| --- | --- |
| `tasks(user_id, status, created_at)` | ユーザの未完了/完了タスク一覧取得、統計の当日タスク集計 |
| `screentime_records(user_id, date)` UNIQUE | 特定日の記録取得・upsert（unique制約がインデックスを兼ねる） |
| `chat_logs(user_id, created_at)` | ユーザのチャットログ一覧取得、日付絞り込み、直近文脈取得 |
| `voice_files(character_id, created_at)` | キャラ別の音声ファイル履歴取得 |

## 11. 初期データ管理方針

- 初期キャラクターデータ（`characters` の4レコード）は、`001_init_master_and_user` migration 内の `INSERT ... ON CONFLICT (id) DO NOTHING` で投入する（既存実装を正本として維持）。
- `users`, `tasks`, `screentime_records`, `chat_logs`, `voice_files` に migration 内で初期データは投入しない。
- 別途 seed script（`prisma`等のORM seedではなく、Goの専用コマンドやSQLスクリプト）を導入する場合は、`characters` の初期データを migration から seed script へ移行できるが、MVP では migration 内 `INSERT` 方式を継続する。

## 12. マイグレーション計画

| migration | テーブル | 状態 |
| --- | --- | --- |
| `000001_init_master_and_user` | `characters`, `users`（+初期キャラデータ投入） | 適用済み |
| `000002_init_task_and_activity` | `tasks` | 適用済み |
| `000003_init_screentime` | `screentime_records` | 未適用（後続 migration 実装で追加） |
| `000004_init_voice_and_chat` | `voice_files`, `chat_logs`（この順） | 未適用（後続 migration 実装で追加） |

後続の migration 実装では、本ドキュメントの型・制約・外部キー・インデックス定義をそのまま `up`/`down` SQL に反映する。

# 二次元秘書AIタスク・娯楽管理アプリ DB設計

## 1. 目的

このドキュメントは、二次元秘書AIタスク・娯楽管理アプリのMVPで使用するDB設計を定義する。

DBは、ユーザ設定の保存、タスク管理、娯楽時間の追跡、および秘書キャラとのチャット履歴の永続化を担う。「社長（ユーザ）の行動を秘書が把握し、適切に反応できる」状態を維持するためのデータ基盤とする。

## 2. 基本方針

MVPでは、開発速度とローカル完結型の動作を優先し、SQLiteを採用する。

方針:
- **マイグレーション管理**: スキーマ変更はマイグレーションファイルで厳格に管理する。
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

## 4. マスターテーブル

### characters

秘書キャラクターの定義を保持する。

| カラム | 型 | 説明 |
| --- | --- | --- |
| id | text | キャラクターID（例: `istj_sec`） |
| name | text | キャラクター名 |
| mbti_type | text | MBTIタイプ（例: `ISTJ`, `ENFJ`） |
| personality_desc | text | 性格説明（AIプロンプト用） |
| speech_style | text | 口調の定義（AIプロンプト用） |
| system_prompt_fragment | text | AI用のシステムプロンプト断片 |
| voice_key | text | 使用するボイスエンジンのキー |
| icon_path | text | アイコン画像のパス |
| standing_image_path | text | 立ち絵画像のパス |
| sample_voice_path | text | サンプルボイスのパス |
| created_at | datetime | 作成日時 |
| updated_at | datetime | 更新日時 |

## 5. ユーザ・設定テーブル

### users

ユーザの基本プロフィールと現在設定を保持する。

| カラム | 型 | 説明 |
| --- | --- | --- |
| id | text | ユーザID |
| name | text | 本名/ユーザ名 |
| president_name | text | キャラクターからの呼び名（例: `社長`） |
| selected_character_id | text | 現在選択中のキャラクターID |
| target_entertainment_minutes | integer | 1日の目標娯楽時間（分） |
| created_at | datetime | 作成日時 |
| updated_at | datetime | 更新日時 |

## 6. タスク・活動テーブル

### tasks

ユーザが登録したタスクを管理する。

| カラム | 型 | 説明 |
| --- | --- | --- |
| id | text | タスクID |
| user_id | text | ユーザID |
| title | text | タスク内容 |
| status | text | 状態（`todo`, `done`） |
| created_at | datetime | 登録日時 |
| completed_at | datetime | 完了日時 |

### entertainment_records

手動入力された娯楽時間を記録する。

| カラム | 型 | 説明 |
| --- | --- | --- |
| id | text | 記録ID |
| user_id | text | ユーザID |
| date | date | 対象日 |
| minutes | integer | 娯楽時間（分） |
| target_minutes_snapshot | integer | その時点の目標時間（統計用） |
| created_at | datetime | 記録日時 |
| updated_at | datetime | 更新日時 |

## 7. 統計・ログテーブル


### chat_logs

ユーザとAIの会話履歴を保持する。

| カラム | 型 | 説明 |
| --- | --- | --- |
| id | text | ログID |
| user_id | text | ユーザID |
| character_id | text | 会話相手のキャラID |
| role | text | 発言者（`user`, `assistant`） |
| message | text | 発言内容 |
| voice_file_id | text | 生成されたボイスファイルのID（存在する場合） |
| created_at | datetime | 発言日時 |

### voice_files

生成された音声ファイルのメタデータを管理する。

| カラム | 型 | 説明 |
| --- | --- | --- |
| id | text | ファイルID（例: `voice_file_001`） |
| character_id | text | 生成に使用したキャラID |
| text | text | 読み上げ元テキスト |
| file_path | text | ローカルの保存パス |
| created_at | datetime | 生成日時 |

## 8. インデックス方針

効率的なデータ取得のため、以下のインデックスを付与する。

- `tasks(user_id, status, created_at)`: 未完了タスクの取得用
- `entertainment_records(user_id, date)`: 特定日の記録確認用
- `chat_logs(user_id, created_at)`: 直近の文脈取得用
- `voice_files(character_id, created_at)`: 音声ファイルの履歴取得用

## 9. マイグレーション計画

1. **001_init_master_and_user.sql**: `characters`, `users` の作成と初期キャラデータの投入。
2. **002_init_task_and_activity.sql**: `tasks`, `entertainment_records` の作成。
3. **003_init_logs.sql**: `chat_logs`, `voice_files` の作成。

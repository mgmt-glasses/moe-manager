# DB 方針

このドキュメントは、既存の `detail/db-design.md` をもとに DB 設計方針を整理したものです。

## 目的

DB は、ユーザ設定、タスク、娯楽時間、チャット履歴を永続化し、秘書キャラがユーザの行動を把握して反応できる状態を支えます。

## 基本方針

- MVP では PostgreSQL を採用する。
- 開発速度と本番環境との一貫性を優先する。ローカル開発では Docker で PostgreSQL を起動する。
- スキーマ変更はマイグレーションファイルで管理する。
- キャラクター定義などのマスターデータと、タスクやログなどのユーザデータを分ける。
- MVP では専用の統計テーブルを持たず、タスクや娯楽時間の記録から都度計算する。
- チャットログは、AI が参照しやすい粒度で保存する。

`detail/仕様書.md` には統計データを保存対象として扱う記述がありますが、現時点の DB 方針では `daily_stats` のような専用統計テーブルは作りません。
必要になった時点で追加マイグレーションを作成します。

## データ分類

| 分類 | 例 | 方針 |
| --- | --- | --- |
| ユーザ設定 | ユーザ名、呼び名、目標娯楽時間 | アプリの基本動作を決定する |
| キャラクター定義 | MBTI、性格、口調、ボイス設定 | アプリ同梱の静的マスターデータ |
| タスク管理 | タスク内容、状態、作成/完了日時 | ユーザの主活動を記録する |
| 活動記録 | 日別の娯楽時間 | 統計の元データにする |
| チャットログ | 会話内容、発言者、日時 | 過去文脈の保持に使う |
| 音声ファイル | 生成テキスト、保存パス | TTS 結果の参照に使う |

## 主要テーブル

| テーブル | 役割 |
| --- | --- |
| `characters` | 秘書キャラクターの定義 |
| `users` | ユーザプロフィールと現在設定 |
| `tasks` | ユーザが登録したタスク |
| `screentime_records` | 日付ごとの娯楽時間 |
| `chat_logs` | ユーザと AI の会話履歴 |
| `voice_files` | 生成音声ファイルのメタデータ |

命名整理:

- DB テーブル名は現行実装に合わせて `screentime_records` を正とする。
- API パスはフロントエンドの意味が分かりやすい `entertainment-records` を使ってよい。
- 既存ドキュメント内の `entertainment_records` と `entertainment_times` は旧表記として扱う。
- キャラクターのボイス設定は `voice_preset_id` を正とし、既存 DB 設計書の `voice_key` は旧表記として扱う。

## インデックス方針

- `tasks(user_id, status, created_at)`: 未完了タスク取得用
- `screentime_records(user_id, date)`: 特定日の記録取得用
- `chat_logs(user_id, created_at)`: 直近文脈取得用
- `voice_files(character_id, created_at)`: 音声履歴取得用

## マイグレーション計画

初期マイグレーションは以下の粒度を基本とします。

1. `001_init_master_and_user.sql`
   - `characters`, `users` の作成
   - 初期キャラデータの投入
2. `002_init_task_and_activity.sql`
   - `tasks`, `screentime_records` の作成
3. `003_init_logs.sql`
   - `chat_logs`, `voice_files` の作成

## マイグレーション実行方法

MVP では外部ツール（Alembic 等）を使わず、アプリ起動時に SQL ファイルを順番に適用するシンプルな方式を基本とします。

```python
# apps/gateway/gateway/migrations.py（例）
import pathlib
import psycopg

MIGRATIONS_DIR = pathlib.Path("migrations")

def apply_migrations(dsn: str) -> None:
    with psycopg.connect(dsn) as conn:
        applied = _get_applied(conn)
        for sql_file in sorted(MIGRATIONS_DIR.glob("*.sql")):
            if sql_file.name not in applied:
                conn.execute(sql_file.read_text())
                _record_applied(conn, sql_file.name)
        conn.commit()
```

将来的にスキーマ変更が増えてきたら Alembic の導入を検討します。

## 将来の DB 移行

将来的には Supabase（マネージド PostgreSQL）への移行を想定します。
同じ PostgreSQL のため接続先 URL の変更で移行でき、
domain と use case は repository port にのみ依存させ、DB 実装の差し替えができる状態を保ちます。

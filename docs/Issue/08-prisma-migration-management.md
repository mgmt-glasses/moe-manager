# Prisma で DB マイグレーションを管理できる

## 目的

プロダクト全体で未完成の DB スキーマを、Prisma を使って一貫して管理できる状態にする。
各機能 Issue が個別に DB 定義や SQL を持つのではなく、共通の Prisma schema と migration 履歴を正として、PostgreSQL の初期構築、差分適用、開発環境での再現を行えるようにする。

## 背景

既存ドキュメントでは MVP の DB として PostgreSQL を採用し、`characters`, `users`, `tasks`, `screentime_records`, `chat_logs`, `voice_files` を主要テーブルとして定義している。
一方で、現状の実装は Python の SQLite repository が中心で、リポジトリには Prisma schema や migration 実行手順が存在しない。

今後の機能 Issue で DB を扱う前に、DB スキーマの正本、マイグレーション作成方法、適用方法、検証方法を決めておく必要がある。

## 対象範囲

### やること

- Prisma を使用した DB スキーマ管理の配置と運用方針を決める。
- `docs/detail/db-design.md` と `docs/database-guidelines.md` をもとに、MVP で必要なテーブルを Prisma schema に整理する。
- PostgreSQL 向けの初期 migration を作成できる状態にする。
- ローカル開発で migration を適用できるコマンドを定義する。
- migration 適用後の DB 状態を確認する手順を定義する。
- Prisma を schema/migration 管理に使う範囲を明確にする。
- 既存 Issue と方針ドキュメントの DB 変更方針を Prisma migration 前提に揃える。

### やらないこと

- 各機能 API の実装はこの Issue では扱わない。
- 既存 Python SQLite repository の完全置き換えはこの Issue では扱わない。
- 本番 DB の作成、ホスティング、バックアップ運用はこの Issue では扱わない。
- 認証・認可用の DB 設計はこの Issue では扱わない。
- 統計専用テーブルは作成しない。統計は既存方針どおり `tasks` と `screentime_records` から都度計算する。
- Prisma Client を使った repository adapter 実装はこの Issue では扱わない。
- キャラクターの長期記憶、会話要約、好感度、関係性 state、ユーザプロファイル抽出はこの Issue では扱わない。

## 実装方針

- 対象モジュール: `db`, `docs`
- 想定ブランチ: `feature/db-prisma-migrations`
- 主な変更ファイル/ディレクトリ:
  - `prisma/schema.prisma`
  - `prisma/migrations/`
  - `prisma/seed.*` または同等の seed script
  - `.env.example`
  - `package.json` または同等の task 定義ファイル
  - `docs/database-guidelines.md`
  - `docs/implementation-guide.md`
  - `docs/detail/db-design.md`
  - `docs/Issue/README.md`
- 依存する Issue: なし
- 後続 Issue: `01-character-selection.md`, `04-task-management.md`, `05-screentime-recording.md`, `07-chat-logs.md`

DB は PostgreSQL を正とし、接続情報は `DATABASE_URL` で渡す。
Prisma は schema と migration 履歴の管理に限定し、Prisma Client の導入は後続の DB adapter 実装 Issue で判断する。
domain/use case から Prisma Client を直接参照しない方針は、この Issue でドキュメントに明記する。

## 要件

### スキーマ管理

- [ ] `characters`, `users`, `tasks`, `screentime_records`, `chat_logs`, `voice_files` を Prisma model として定義する。
- [ ] `docs/database-guidelines.md` の命名整理に従い、娯楽時間の DB テーブル名は `screentime_records` とする。
- [ ] キャラクターのボイス設定は DB 上で `voice_preset_id` として扱う。
- [ ] 日時は PostgreSQL の `timestamptz` 相当で扱う。
- [ ] 日付のみを扱う `screentime_records.date` は PostgreSQL の `date` 相当で扱う。
- [ ] 主要な外部キー関係を Prisma schema 上に定義する。
- [ ] 既存 DB 設計にあるインデックス方針を Prisma schema 上に反映する。
- [ ] カラム定義は `docs/detail/db-design.md` を正とし、既存方針で補正済みの命名を優先する。
- [ ] `screentime_records(user_id, date)` は同一ユーザ・同一日の重複を避けるため unique 制約を付ける。
- [ ] MVP のタスク削除は物理削除とし、初期 schema には `deleted_at` を含めない。
- [ ] `character_memories`, `relationship_states`, `user_profiles` などの長期記憶・関係性系テーブルは初期 schema に含めない。

### マイグレーション管理

- [ ] 初期 migration で MVP に必要な全テーブルを作成できる。
- [ ] migration は `prisma/migrations/` に履歴として保存する。
- [ ] 開発環境で `prisma migrate dev` 相当のコマンドを実行できる。
- [ ] CI または検証用に `prisma migrate deploy` 相当のコマンドを実行できる。
- [ ] schema と migration の差分が残っていないことを確認できる。
- [ ] DB を破壊的に初期化する操作は通常の migration 適用手順と分けて扱う。
- [ ] 初期キャラクターデータは migration ではなく seed script で投入する。

### 開発環境

- [ ] ローカル PostgreSQL への接続は `DATABASE_URL` で設定する。
- [ ] `.env.example` に必要な環境変数を記載する。
- [ ] migration 実行前に PostgreSQL を用意する手順をドキュメント化する。
- [ ] Prisma CLI の実行方法を README または DB 方針に記載する。
- [ ] Docker で PostgreSQL を起動する場合の手順を記載する。Docker Compose ファイルをリポジトリに含めるか、`docker run` の手順に留めるかは実装時に決める。

### アプリケーション境界

- [ ] domain/use case は Prisma Client に依存しない。
- [ ] Prisma Client 導入や repository adapter 実装は後続 Issue のスコープとする。

## 受け入れ条件

- [ ] `prisma/schema.prisma` に MVP DB の主要テーブルが定義されている。
- [ ] 初期 migration を空の PostgreSQL DB に適用できる。
- [ ] migration 適用後、主要テーブルとインデックスが作成されていることを確認できる。
- [ ] `DATABASE_URL` を使った migration 実行手順がドキュメントに記載されている。
- [ ] `docs/database-guidelines.md` のマイグレーション方針が Prisma 前提に更新されている。
- [ ] 既存 Issue から参照される DB 変更方針が Prisma migration と矛盾しない。
- [ ] domain/use case から Prisma Client を直接参照しない方針が明記されている。
- [ ] 初期キャラクターデータを seed で投入できる。
- [ ] 長期記憶・関係性系テーブルを初期 schema に含めないことが明記されている。

## API / DB 変更

### API

- この Issue では API エンドポイントを追加しない。

### DB

- `characters`
- `users`
- `tasks`
- `screentime_records`
- `chat_logs`
- `voice_files`

初期 migration では上記テーブル、外部キー、インデックスを作成する。
初期キャラクターデータは migration ではなく seed script で投入する。
長期記憶・関係性 state・ユーザプロファイル抽出用のテーブルは、この Issue の初期 DB には含めない。

## テスト・動作確認

- [ ] 空の PostgreSQL DB に migration を適用する。
- [ ] migration 適用後に Prisma schema と DB の差分がないことを確認する。
- [ ] 主要テーブル、外部キー、インデックスの存在を確認する。
- [ ] `screentime_records(user_id, date)` の unique 制約を確認する。
- [ ] seed 実行後に初期キャラクターデータが投入されていることを確認する。
- [ ] `DATABASE_URL` 未設定時に実行手順上のエラー原因が分かる。
- [ ] DB 方針ドキュメントと Issue ドキュメントの記述が矛盾していない。

## 未決事項

- Docker Compose ファイルをリポジトリに含めるか、`docker run` の手順に留めるか。
- 既存 Python SQLite repository をどの Issue で置き換えるか。
- キャラクターの長期記憶・関係性 state・ユーザプロファイル抽出を扱う別 Issue をいつ作成するか。

## 参照ドキュメント

- `docs/database-guidelines.md`
- `docs/implementation-guide.md`
- `docs/detail/db-design.md`
- `docs/api-guidelines.md`

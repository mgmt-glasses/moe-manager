# CLAUDE.md

このファイルは、リポジトリ内のコードを操作する際に Claude Code (claude.ai/code) へ提供するガイダンスです。

## コマンド

このプロジェクトはパッケージマネージャーに [uv](https://docs.astral.sh/uv/) を使用し、ワークスペース構成を採用している。

```bash
# ワークスペース全体の依存関係をインストール
uv sync

# ゲートウェイサーバーを起動（開発用）
uv run --package gateway uvicorn gateway.main:app --reload

# 特定パッケージのテストを実行
uv run --package moe-task pytest packages/task/

# テストファイルを1つだけ実行
uv run --package moe-task pytest packages/task/tests/test_use_cases.py

# パッケージに依存関係を追加
uv add --package moe-task <パッケージ名>
```

## アーキテクチャ

**モジュラーモノリス × ヘキサゴナルアーキテクチャ（Ports & Adapters）**

単一リポジトリ・単一デプロイ。各ビジネスドメインは `packages/` 配下に独立した Python パッケージとして存在し、`apps/gateway/` でのみ統合される。

### パッケージとドメインの対応

| パッケージ（`packages/`） | Python モジュール | 責務 |
|---|---|---|
| `user/` | `moe_user` | ユーザー登録・設定・キャラ選択保存 |
| `character/` | `moe_character` | MBTIキャラ一覧・詳細・ボイスサンプル |
| `task/` | `moe_task` | タスクCRUD・完了率計算 |
| `screentime/` | `moe_screentime` | 娯楽時間入力・目標差分計算 |
| `statistics/` | `moe_statistics` | 今日/週次サマリー集計 |
| `voice-library/` | `mkh_voice` | AIチャット（LLM）+ TTS音声生成 |

`moe_statistics` のみ、他パッケージのテーブル（tasks + screentime_records）を SQL JOIN で横断クエリすることが許可されている。これは意図的な設計。

### 各パッケージのヘキサゴナル層

```
domain/models.py      ← Pydantic エンティティ。外部依存ゼロ（標準ライブラリ + pydantic のみ）
domain/ports.py       ← Protocol によるインターフェース定義（アウトバウンドポート）
domain/use_cases.py   ← ビジネスロジック。models と ports にのみ依存
adapters/inbound/api/ ← FastAPI ルーター。HTTP ↔ ドメインオブジェクトの変換
adapters/outbound/repositories/ ← ポートの SQLite/DB 実装
```

**依存の方向ルール**: 矢印は常に内側を向く。`domain` は `adapters` を import しない。`adapters/outbound` は `adapters/inbound` を import しない。パッケージ同士は相互 import 禁止。モジュール間のデータ受け渡しはプリミティブ型（`str`, `int`, `date`）で行う。

### リクエストの流れ

```
モバイルアプリ → POST /tasks
  → apps/gateway/main.py          （ルーティングのみ、ロジックなし）
  → moe_task/.../task_router.py   （HTTP → ドメインオブジェクトに変換）
  → moe_task/domain/use_cases.py  （ビジネスロジック）
  → sqlite_task_repository.py     （Port Protocol 経由で永続化）
```

### モジュール間の連携（チャットフロー）

`mkh_voice` のチャット機能は、gateway 層で全モジュールを統合する：
1. `moe_user` → 現在の `character_id` を取得
2. `moe_character` → `voice_preset_id` を取得
3. `moe_task` + `moe_screentime` → 今日の状況を取得
4. `mkh_voice` → キャラ人格 + 状況をプロンプトに乗せて LLM 応答生成 → `voice_preset_id` を使って TTS 音声生成

### DB

開発環境は全パッケージ共通の `data/moe.db` SQLite ファイルを使用。アウトバウンドリポジトリは SQLAlchemy を使っているため、PostgreSQL（Supabase）への移行は接続 URL の変更だけで完了し、domain/use_cases の変更は不要。

## 開発ルール

### 新機能の実装順序

1. `domain/models.py` — エンティティを定義
2. `domain/ports.py` — Protocol インターフェースを定義
3. `domain/use_cases.py` — ビジネスロジックを実装
4. `adapters/outbound/repositories/` — SQLite アダプタを実装
5. `adapters/inbound/api/` — FastAPI ルーターを実装
6. `apps/gateway/main.py` — `include_router(...)` を追加

### ブランチ・コミット規則

ブランチ名: `feature/<モジュール>-<機能名>`、`fix/<モジュール>-<修正内容>`、`hotfix/<修正内容>`  
`develop` から切り、PR は `develop` へ。マージは Squash Merge を基本とする。

コミットは [Conventional Commits](https://www.conventionalcommits.org/) に準拠：
```
feat(task): タスク管理のCRUD APIを追加
fix(screentime): 日付計算が1日ずれる問題を修正
```

有効なスコープ: `user`, `character`, `task`, `screentime`, `statistics`, `voice`, `gateway`

### PR ルール

- 原則 1 PR = 1 モジュール（複数モジュールにまたがる場合は PR の概要に理由を明記）
- 変更行数の目安は 300 行以内。超える場合は分割を検討する
- レビューコメントのプレフィックス: `MUST`（マージ前に必須対応）、`IMO`（提案・任意）、`NIT`（細かい指摘・任意）、`Q`（質問）

# Claude Code 向け指示

## Claude Codeを使って実装行う際
`docs/Rule.md` を必ず読み、実装して欲しい
# Issue ドキュメント

このディレクトリは、Issue ごとの作業範囲、受け入れ条件、実装観点を整理するための場所です。
PRD/TRD の詳細を再掲せず、実装ブランチを切る前に確認する作業単位として使います。

## 運用方針

- 1 Issue は原則 1 PR に対応させる。
- 1 PR は原則 1 モジュール内の変更に閉じる。
- 複数モジュールにまたがる場合は、Issue 内に理由、影響範囲、依存 Issue を明記する。
- ブランチは `develop` から切り、`feature/<module>-<feature>` 形式にする。
- 実装中に仕様や設計判断が変わった場合は、関連する方針ドキュメントも同じ PR で更新する。

## Go 実装方針

- MVP バックエンドの使用言語は Go とする。
- 実行エントリポイントと依存性注入は `cmd/api/` に置く。
- ドメインごとのモデル、ユースケース、interface、adapter は `internal/<domain>/` に置く。
- ドメイン間で実装型を直接参照せず、`user_id` や集計値など必要最小限の値を渡す。
- DB、LLM、TTS、ファイルストレージは interface の背後に置き、テストでは fake に差し替える。
- 既存の Python 実装はプロトタイプとして参照してよいが、Go の domain から直接 import・実行しない。
- Python 製 TTS を継続利用する場合は、HTTP 等の明示的な外部サービス境界を設ける。

基本配置:

```txt
cmd/api/                    # HTTP サーバー起動、ルーティング、依存性注入
internal/<domain>/          # domain / use case / interface
internal/<domain>/adapter/  # PostgreSQL、外部 API 等の実装
migrations/                 # PostgreSQL マイグレーション
```

## Issue 一覧

| ファイル | 対象機能 |
| --- | --- |
| `01-character-selection.md` | ユーザが MBTI 別の秘書キャラを選べる |
| `02-character-chat.md` | 選んだ秘書キャラとチャットできる |
| `03-character-voice.md` | キャラの返答にボイスが付く |
| `04-task-management.md` | タスクを登録・完了できる |
| `05-screentime-recording.md` | スクリーンタイム、または娯楽時間を記録できる |
| `06-statistics-summary.md` | タスク達成状況と娯楽時間を統計として確認できる |
| `07-chat-logs.md` | チャットログを確認できる |
| `08-postgresql-schema-design.md` | MVP の PostgreSQL スキーマを設計できる |
| `10-vertex-ai-context-caching.md` | Vertex AI Context Caching で LLM トークンコストを削減する |

## テンプレート

```md
# <機能名>

## 目的

## 背景

## 対象範囲

### やること

### やらないこと

## 実装方針

- 対象モジュール:
- 想定ブランチ:
- 主な変更ファイル/ディレクトリ:
- 依存する Issue:
- 後続 Issue:

## 受け入れ条件

- [ ] 

## API / DB 変更

### API

### DB

## テスト・動作確認

- [ ] 

## 参照ドキュメント
```

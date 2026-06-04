# 実装ガイド

このドキュメントは、実装時の進め方と配置場所をまとめます。

## 実装順序

新しい機能は、原則として以下の順で実装します。

1. `internal/<domain>/model.go` にエンティティ、値オブジェクトを定義する。
2. 利用側パッケージに DB や外部 API の interface を定義する。
3. `internal/<domain>/service.go` にビジネスロジックを実装する。
4. `internal/<domain>/adapter/` に DB や外部 API の実装を追加する。
5. `internal/<domain>/handler.go` に HTTP handler を追加する。
6. `cmd/api/` でルーティングと依存性注入を行う。

domain から先に作ることで、外部依存なしでビジネスロジックを確認しやすくします。

## 追加するものと配置場所

| 追加するもの | 配置場所 |
| --- | --- |
| ドメインモデル | `internal/<domain>/model.go` |
| 抽象インターフェース | `internal/<domain>/*.go` の利用側パッケージ |
| ビジネスロジック | `internal/<domain>/service.go` |
| REST API | `internal/<domain>/handler.go` |
| DB アクセス | `internal/<domain>/adapter/` |
| 外部 API 連携 | `internal/<domain>/adapter/` |
| 統合ルーティング | `cmd/api/` |

## モジュール対応

以下はアーキテクチャ上のモジュール対応です。
API 実装順やリリース順の Phase とは別物として扱います。

| 対応 | モジュール | 主な責務 |
| --- | --- | --- |
| User | `internal/user` | ユーザー登録、初期設定、キャラ選択保存 |
| Character | `internal/character` | MBTI キャラ一覧、詳細、ボイスサンプル |
| Task | `internal/task` | タスク CRUD、完了率計算 |
| ScreenTime | `internal/screentime` | 娯楽時間入力、目標差分計算 |
| Statistics | `internal/statistics` | 今日、週次サマリー集計 |
| Chat | `internal/chat` | AI チャット、チャットログ |
| Voice | `internal/voice` | TTS 音声生成、音声ファイル管理 |

API の優先実装順は [API 方針](api-guidelines.md) を参照します。

## API 実装方針

- Base URL は `/api/v1` を前提にする。
- 画面単位ではなく、ドメイン単位で API を分ける。
- フロントエンドに AI 生成やボイス生成の詳細を漏らさない。
- MVP では認証は簡易実装でもよいが、API 上は `userId` を前提にする。
- 日付は `YYYY-MM-DD`、日時は ISO 8601 を使う。
- HTTP router は Go 標準ライブラリ互換の interface を持つものを使い、handler にビジネスロジックを置かない。

共通レスポンス形式:

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

エラー時:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "入力内容が正しくありません"
  }
}
```

## DB 実装方針

- MVP では PostgreSQL を採用する。
- 全モジュールが同一データベースを参照する。
- マスターデータとユーザデータを分ける。
- 統計は MVP では原則として都度計算する。
- スキーマ変更はマイグレーションファイルで管理する。
- 娯楽時間の DB テーブル名は現行実装に合わせて `screentime_records` を正とする。

接続情報は環境変数 `DATABASE_URL` で渡す:

```txt
postgresql://user:password@localhost:5432/moe
```

## 実装前チェック

作業前に以下を確認します。

- 変更対象のモジュールと API 優先実装順はどれか。
- 変更は 1 モジュールに閉じるか。
- 複数ドメインをまたぐ処理の手順が、処理を所有する `internal/<domain>` に置かれているか。
- `internal/statistics` の集計は Query interface と adapter で扱えているか。
- domain に外部依存が入り込まないか。
- 新しい外部依存に interface が必要か。
- API と DB の命名が既存方針と矛盾しないか。

## PR 前チェック

PR 作成前に以下を確認します。

- 変更目的が PR 概要で説明できる。
- 動作確認内容を書ける。
- `domain` が `adapters` を import していない。
- 他モジュールの内部実装を直接 import していない。
- API レスポンス形式が方針と一致している。
- DB 変更がある場合、マイグレーション方針と整合している。
- 関連ドキュメントを更新している。
- `go test ./...` が成功している。

# 実装ガイド

このドキュメントは、実装時の進め方と配置場所をまとめます。

## 実装順序

新しい機能は、原則として以下の順で実装します。

1. `domain/models.py` にエンティティ、値オブジェクト、Enum を定義する。
2. `domain/ports.py` に外部依存の抽象インターフェースを定義する。
3. `domain/use_cases.py` にビジネスロジックを実装する。
4. `adapters/outbound/` に DB や外部 API の実装を追加する。
5. `adapters/inbound/api/` に FastAPI ルーターを追加する。
6. `apps/gateway/gateway/main.py` にルーターを追加する。

domain から先に作ることで、外部依存なしでビジネスロジックを確認しやすくします。

## 追加するものと配置場所

| 追加するもの | 配置場所 |
| --- | --- |
| ドメインモデル | `packages/<module>/moe_<module>/domain/models.py` |
| 抽象インターフェース | `packages/<module>/moe_<module>/domain/ports.py` |
| ビジネスロジック | `packages/<module>/moe_<module>/domain/use_cases.py` |
| REST API | `packages/<module>/moe_<module>/adapters/inbound/api/` |
| DB アクセス | `packages/<module>/moe_<module>/adapters/outbound/repositories/` |
| 外部 API 連携 | `packages/<module>/moe_<module>/adapters/outbound/` |
| 統合ルーティング | `apps/gateway/gateway/main.py` |

## モジュール対応

以下はアーキテクチャ上のモジュール対応です。
API 実装順やリリース順の Phase とは別物として扱います。

| 対応 | モジュール | 主な責務 |
| --- | --- | --- |
| User | `moe_user` | ユーザー登録、初期設定、キャラ選択保存 |
| Character | `moe_character` | MBTI キャラ一覧、詳細、ボイスサンプル |
| Task | `moe_task` | タスク CRUD、完了率計算 |
| ScreenTime | `moe_screentime` | 娯楽時間入力、目標差分計算 |
| Statistics | `moe_statistics` | 今日、週次サマリー集計 |
| Voice / Chat | `mkh_voice` | AI チャット、TTS 音声生成 |

API の優先実装順は [API 方針](api-guidelines.md) を参照します。

## API 実装方針

- Base URL は `/api/v1` を前提にする。
- 画面単位ではなく、ドメイン単位で API を分ける。
- フロントエンドに AI 生成やボイス生成の詳細を漏らさない。
- MVP では認証は簡易実装でもよいが、API 上は `userId` を前提にする。
- 日付は `YYYY-MM-DD`、日時は ISO 8601 を使う。

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
- 他モジュールとの連携は gateway で扱えるか。
- `moe_statistics` の集計は Query Port と outbound adapter で扱えているか。
- domain に外部依存が入り込まないか。
- 新しい外部依存に Port が必要か。
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

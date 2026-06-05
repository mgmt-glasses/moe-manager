# アーキテクチャ方針

このドキュメントは、Moe Manager の設計判断と依存ルールをまとめます。
MVP バックエンドの使用言語は Go とします。
`detail/ARCHITECTURE.md` に残る Python / FastAPI 構成は旧設計として扱い、移行時の参考に限定します。

## 採用アーキテクチャ

Moe Manager は、モジュラーモノリスとヘキサゴナルアーキテクチャを採用します。

- Go アプリケーションは単一リポジトリ、単一デプロイを前提にする。PostgreSQL、LLM、TTS などの外部サービスはこの単一デプロイに含めない。
- 機能ごとに `internal/` 配下の独立パッケージとして分ける。
- 各パッケージ内では domain を中心に置き、外部依存を adapter に閉じ込める。
- DB、LLM、TTS などの外部依存は interface を通して扱う。

## 全体構成

```txt
moe-manager/
├── cmd/
│   └── api/              # HTTP サーバー起動、ルーティング、依存性注入
├── internal/
│   ├── user/             # ユーザ管理
│   ├── character/        # キャラ管理
│   ├── task/             # タスク管理
│   ├── screentime/       # スクリーンタイム管理
│   ├── statistics/       # 統計管理
│   ├── chat/             # AI チャット、チャットログ
│   └── voice/            # TTS 音声生成、音声ファイル管理
├── prisma/
│   └── migrations/       # Prisma による PostgreSQL マイグレーション
└── go.mod
```

## モジュール内構成

各モジュールは以下の構成を基本とします。

```txt
internal/<module>/
├── model.go              # エンティティ、値オブジェクト
├── service.go            # ユースケース
├── repository.go         # DB 等の interface
├── handler.go            # HTTP handler
└── adapter/
    ├── postgres.go       # PostgreSQL 実装
    └── external.go       # LLM / TTS 等の外部 API 実装
```

ファイルは責務に応じて分割してよく、上記のファイル名へ固定しません。
interface は利用側パッケージに定義し、実装詳細を domain/service から切り離します。

## 依存方向

依存方向は常に外側から内側へ向けます。

```txt
handler / adapter -> service -> interface / model
```

| レイヤー | 依存してよいもの | 依存してはいけないもの |
| --- | --- | --- |
| model / service | Go 標準ライブラリ、自パッケージの interface | HTTP、DB driver、外部 API SDK、他ドメインの実装型 |
| handler | service、API DTO | DB や外部 API の具体実装 |
| adapter | model、interface、外部ライブラリ | handler |
| `cmd/api` | 各パッケージの公開 API、adapter | ビジネスロジック |

## モジュール境界

- モジュール間で内部実装を直接 import しない。
- domain から別モジュールの domain を直接参照しない。
- `user_id` や `character_id` などの識別子は primitive として受け渡す。
- 複数ドメインをまたぐユースケースは、その処理を所有する `internal/<domain>` に置き、依存先を interface として定義する。
- `cmd/api` は依存性注入に徹し、ビジネスロジックを持たない。
- 共通型が必要になった場合のみ `internal/shared` などの共通パッケージを検討する。

例外として、`internal/statistics` は集計のために `tasks` と `screentime_records` を読む必要があります。
この場合も他モジュールの実装型を直接参照せず、`internal/statistics` 側の Query interface と adapter で集計クエリを扱います。
`cmd/api` は統計 API のルーティングと依存性注入を担当し、集計ロジックは持ちません。

実装パターン:

```go
type StatisticsQuery interface {
	GetDailyStats(ctx context.Context, userID string, date time.Time) (DailyStats, error)
}
```

`internal/statistics` の Query adapter は `tasks` / `screentime_records` テーブルを直接クエリします。
他モジュールの実装型は import せず、テーブル名だけを知ります。

## `cmd/api` の責務

`cmd/api` は、各モジュールを束ねる統合エントリポイントです。

- ルーティングと依存性注入を担当する。
- ビジネスロジックを持たない。
- 複数モジュールをまたぐ処理に必要な interface と adapter を組み合わせる。
- モジュール内の repository 実装詳細を他モジュールへ漏らさない。

## チャット・TTS との関係

- チャット生成とログ管理は `internal/chat` が担う。
- 音声生成と音声ファイル管理は `internal/voice` が担う。
- LLM と TTS は interface の背後に置く。
- Python 製 TTS を利用する場合は別サービスとして起動し、Go adapter から HTTP 等で接続する。
- domain 内部では `voice_preset_id` を正とし、API では `voiceId` として返してよい。
- DB 設計書の `voice_key` は旧表記として扱う。

チャット時は、ユーザ設定、キャラ設定、タスク状況、娯楽時間、直近ログを組み合わせて、LLM 応答と TTS 音声生成につなげます。

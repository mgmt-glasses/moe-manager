# アーキテクチャ方針

このドキュメントは、Moe Manager の設計判断と依存ルールをまとめます。
詳細なディレクトリ構成は既存の `detail/ARCHITECTURE.md` も参照してください。

## 採用アーキテクチャ

Moe Manager は、モジュラーモノリスとヘキサゴナルアーキテクチャを採用します。

- 単一リポジトリ、単一デプロイを前提にする。
- 機能ごとに `packages/` 配下の独立パッケージとして分ける。
- 各パッケージ内では domain を中心に置き、外部依存を adapter に閉じ込める。
- DB、LLM、TTS などの外部依存は Port を通して扱う。

## 全体構成

```txt
moe-manager/
├── apps/
│   └── gateway/          # 各モジュールを束ねる FastAPI アプリケーション
├── packages/
│   ├── user/             # ユーザ管理
│   ├── character/        # キャラ管理
│   ├── task/             # タスク管理
│   ├── screentime/       # スクリーンタイム管理
│   ├── statistics/       # 統計管理
│   └── voice-library/    # ボイス・チャット（パッケージ名は mkh_voice）
└── pyproject.toml        # uv workspace 定義
```

## モジュール内構成

各モジュールは以下の構成を基本とします。

```txt
packages/<module>/moe_<module>/
├── domain/
│   ├── models.py         # エンティティ、値オブジェクト、Enum
│   ├── ports.py          # Protocol による抽象インターフェース
│   └── use_cases.py      # ビジネスロジック
└── adapters/
    ├── inbound/
    │   └── api/
    │       └── <module>_router.py
    └── outbound/
        └── repositories/
            └── postgres_<module>_repository.py
```

`voice-library` のみパッケージ名が `mkh_voice` となっており、他モジュールの命名規則（`moe_<module>`）と異なります。

```txt
packages/voice-library/mkh_voice/
├── domain/               # models / ports / use_cases / chat_use_cases
└── adapters/
    ├── inbound/api/      # chat_router
    └── outbound/         # llm / repositories / tts_engine
```

## 依存方向

依存方向は常に外側から内側へ向けます。

```txt
adapter -> use_cases -> ports / models
```

| レイヤー | 依存してよいもの | 依存してはいけないもの |
| --- | --- | --- |
| `domain/models.py` | 標準ライブラリ、Pydantic | ports, use_cases, adapters, 他モジュール |
| `domain/ports.py` | models | use_cases, adapters, 他モジュール |
| `domain/use_cases.py` | models, ports | adapters, 他モジュール |
| `adapters/inbound/` | domain | outbound adapters |
| `adapters/outbound/` | domain/models, domain/ports | inbound adapters, use_cases |

## モジュール境界

- モジュール間で内部実装を直接 import しない。
- domain から別モジュールの domain を直接参照しない。
- `user_id` や `character_id` などの識別子は primitive として受け渡す。
- モジュール間の組み合わせは gateway 層で行う。
- 共通型が必要になった場合は、`packages/shared` などの共通パッケージを検討する。

例外として、`moe_statistics` は集計のために `tasks` と `screentime_records` を読む必要があります。
この場合も他モジュールの domain や use case を直接 import せず、`moe_statistics` 側の Query Port と outbound adapter で集計クエリを扱います。
gateway は統計 API のルーティングと依存性注入を担当し、集計ロジックは持ちません。

実装パターン:

```python
# moe_statistics/domain/ports.py
class StatisticsQueryPort(Protocol):
    def get_daily_stats(self, user_id: str, date: date) -> DailyStats: ...

# moe_statistics/adapters/outbound/repositories/postgres_statistics_repository.py
class PostgresStatisticsRepository:
    def get_daily_stats(self, user_id: str, date: date) -> DailyStats:
        # tasks と screentime_records を直接 JOIN してよい（同一データベース内）
        ...
```

`moe_statistics` の outbound adapter は `tasks` / `screentime_records` テーブルを直接クエリします。
他モジュールのクラスは import せず、テーブル名だけを知ります。

## gateway の責務

`apps/gateway` は、各モジュールの API ルーターを束ねる統合エントリポイントです。

- ルーティングと依存性注入を担当する。
- ビジネスロジックを持たない。
- 複数モジュールをまたぐ処理は、gateway 層で組み合わせる。
- モジュール内の repository 実装詳細を他モジュールへ漏らさない。

## voice-library との関係

`voice-library` は、チャットと音声生成を担う既存ライブラリです。

- `mkh_voice.domain.use_cases.VoiceGenerationUseCase` が音声生成の入口。
- `mkh_voice.domain.chat_use_cases.ChatUseCase` が対話・心情推論の入口。
- domain 内部では `voice_preset_id` を voice-library の `VoicePreset.id` と対応させる。
- API ではフロントエンド向けに `voiceId` として返してよい。`voice_preset_id` → `voiceId` の変換は `adapters/inbound/api/` で行う。
- DB 設計書の `voice_key` は旧表記として扱い、実装時は `voice_preset_id` に寄せる。

チャット時は、ユーザ設定、キャラ設定、タスク状況、娯楽時間、直近ログを組み合わせて、LLM 応答と TTS 音声生成につなげます。

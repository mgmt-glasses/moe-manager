# Moe Manager — アーキテクチャ解説

## 1. 採用アーキテクチャ

**モジュラーモノリス × ヘキサゴナルアーキテクチャ（Ports & Adapters）**

単一リポジトリ・単一デプロイ（モノリス）でありながら、機能ごとに独立したパッケージとして分割する構成。
各パッケージの内部はヘキサゴナルアーキテクチャで実装し、外部依存（DB・LLM・TTS）をPortで隠蔽する。

---

## 2. ディレクトリ構成

```
moe-manager/
├── packages/                       # 機能パッケージ群（ビジネスロジック）
│   ├── user/                       # Phase 1 ユーザー管理
│   │   └── moe_user/
│   │       ├── domain/
│   │       │   ├── models.py       ← エンティティ（User）
│   │       │   ├── ports.py        ← Outbound Port（UserRepositoryPort）
│   │       │   └── use_cases.py    ← ユースケース（UserUseCase）
│   │       └── adapters/
│   │           ├── inbound/api/
│   │           │   └── user_router.py          ← FastAPI ルーター
│   │           └── outbound/repositories/
│   │               └── sqlite_user_repository.py
│   │
│   ├── character/                  # Phase 2 MBTIキャラ選択
│   │   └── moe_character/
│   │       ├── domain/
│   │       │   ├── models.py       ← Character, MBTIType
│   │       │   ├── ports.py        ← CharacterRepositoryPort
│   │       │   └── use_cases.py
│   │       └── adapters/
│   │           ├── inbound/api/character_router.py
│   │           └── outbound/repositories/sqlite_character_repository.py
│   │
│   ├── task/                       # Phase 3 タスク管理
│   │   └── moe_task/
│   │       ├── domain/
│   │       │   ├── models.py       ← Task, TaskStatus
│   │       │   ├── ports.py        ← TaskRepositoryPort
│   │       │   └── use_cases.py
│   │       └── adapters/
│   │           ├── inbound/api/task_router.py
│   │           └── outbound/repositories/sqlite_task_repository.py
│   │
│   ├── screentime/                 # Phase 4 娯楽時間管理
│   │   └── moe_screentime/
│   │       ├── domain/
│   │       │   ├── models.py       ← ScreenTimeRecord
│   │       │   ├── ports.py        ← ScreenTimeRepositoryPort
│   │       │   └── use_cases.py
│   │       └── adapters/
│   │           ├── inbound/api/screentime_router.py
│   │           └── outbound/repositories/sqlite_screentime_repository.py
│   │
│   ├── statistics/                 # Phase 5 統計管理
│   │   └── moe_statistics/
│   │       ├── domain/
│   │       │   ├── models.py       ← DailySummary
│   │       │   ├── ports.py        ← StatisticsQueryPort
│   │       │   └── use_cases.py
│   │       └── adapters/
│   │           ├── inbound/api/statistics_router.py
│   │           └── outbound/repositories/sqlite_statistics_repository.py
│   │
│   └── voice-library/              # Phase 6-7 チャット・ボイス（既存）
│       └── mkh_voice/
│           ├── domain/             ← models / ports / use_cases / chat_use_cases
│           └── adapters/
│               ├── inbound/api/    ← chat_router
│               └── outbound/      ← llm / repositories / tts_engine
│
├── apps/
│   └── gateway/                   # 統合APIサーバー（唯一の起動エントリポイント）
│       └── gateway/
│           └── main.py            ← 全ルーターをinclude_routerで束ねるFastApp
│
├── pyproject.toml                  # uv workspace 定義
├── 仕様書.md
└── ARCHITECTURE.md                 ← 本ファイル
```

---

## 3. ヘキサゴナルアーキテクチャの構造

各パッケージは以下の3層で構成される。

```
┌──────────────────────────────────────────────┐
│                   domain/                     │  ← ビジネスルールの中心
│   models.py   ports.py   use_cases.py         │  ← 外部依存ゼロ
└──────────────────┬───────────────────────────┘
                   │ Port（Protocol）経由のみ
      ┌────────────┴─────────────────┐
      │                              │
  inbound/                       outbound/
  FastAPI Router                 SQLite / LLM / TTS
  （外からの入口）               （外への出口）
```

- **domain** は Python 標準ライブラリと Pydantic のみに依存する
- **inbound adapter**（FastAPI Router）は domain の use_cases を呼ぶ
- **outbound adapter**（SQLite Repository など）は domain の ports を実装する
- use_cases は ports の Protocol 型だけを知り、具体的なアダプタを知らない

---

## 4. モジュールと仕様書 Phase の対応

| Package           | 仕様書 Phase | 主な責務                         |
|-------------------|-------------|----------------------------------|
| `moe_user`        | Phase 1     | ユーザー登録・初期設定・キャラ選択保存 |
| `moe_character`   | Phase 2     | MBTIキャラ一覧・詳細・ボイスサンプル |
| `moe_task`        | Phase 3     | タスクCRUD・完了率計算           |
| `moe_screentime`  | Phase 4     | 娯楽時間入力・目標差分計算       |
| `moe_statistics`  | Phase 5     | 今日/週次サマリー集計            |
| `mkh_voice`       | Phase 6-7   | AIチャット・TTS音声生成（既存）  |

---

## 5. リクエストの流れ（例：タスク登録）

```
モバイルアプリ
    │ POST /tasks
    ▼
apps/gateway/main.py        ← ルーティングのみ、ロジックなし
    │
    ▼
moe_task/adapters/inbound/api/task_router.py  ← HTTPリクエストをドメインオブジェクトに変換
    │
    ▼
moe_task/domain/use_cases.py (TaskUseCase)    ← ビジネスロジック
    │  TaskRepositoryPort 経由
    ▼
moe_task/adapters/outbound/repositories/sqlite_task_repository.py  ← SQLite に永続化
```

---

## 6. モジュール間依存ルール

```
gateway → 全パッケージ（import）
statistics → task / screentime（集計クエリのみ）
その他パッケージ → 相互依存禁止
```

- `moe_task` は `moe_user` を import しない。user_id は単なる String として扱う
- モジュール間でオブジェクトを受け渡す場合は primitive（str, int, date）を使う
- 依存が必要な場合は gateway 層で DI（依存性注入）して組み合わせる

---

## 7. DB 戦略

### 現在：SQLite

```
data/moe.db  ← 全モジュールが同一ファイルを参照
```

全テーブルが同一 SQLite ファイルに作られるため、`moe_statistics` は
`tasks` / `screentime_records` テーブルを JOIN して集計できる。

### 将来：PostgreSQL（Supabase）への移行

SQLAlchemy の接続 URL を変えるだけで移行可能。

```python
# SQLite（開発）
create_engine("sqlite:///data/moe.db")

# PostgreSQL（本番）
create_engine("postgresql://user:pass@host/dbname")
```

各 `SQLiteXxxRepository` を `PostgreSQLXxxRepository` に差し替えるか、
接続URLだけ変えて同じファイルを使い続けることも可能。
Port が抽象化しているため domain/use_cases には変更不要。

---

## 8. voice-library との接続

`character.voice_preset_id` が `mkh_voice` の `VoicePreset.id` と対応する。

```
Character.voice_preset_id → VoicePreset（voice-library）
```

チャット時の流れ：
1. `moe_user` → 選択中の `character_id` を取得
2. `moe_character` → `voice_preset_id` を取得
3. `moe_task` / `moe_screentime` → 今日の状況を取得
4. `mkh_voice` → キャラ人格 + 状況をプロンプトに乗せてLLM応答生成
5. `mkh_voice` → `voice_preset_id` を使ってTTS音声生成

---

## 9. 開発の進め方

Phase 順に以下を繰り返す：

1. `domain/models.py` にエンティティを定義
2. `domain/ports.py` にインターフェースを定義
3. `domain/use_cases.py` にビジネスロジックを実装
4. `outbound/repositories/` に SQLite アダプタを実装
5. `inbound/api/` に FastAPI ルーターを実装
6. `apps/gateway/main.py` にルーターを追加

domain → outbound → inbound の順に実装することで、
ビジネスロジックが最初にテスト可能な状態になる。

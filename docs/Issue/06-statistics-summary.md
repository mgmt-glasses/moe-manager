# タスク達成状況と娯楽時間を統計として確認できる

## 目的

ユーザが今日と過去 7 日分のタスク達成状況、娯楽時間、日次サマリーを確認できるようにする。
統計は専用集計テーブルを持たず、既存データから都度計算する。

## 背景

MVP では、タスクと娯楽時間の状況を可視化し、チャットでの秘書キャラの反応にも利用できるようにする。
`internal/statistics` は例外的に他ドメインが管理するテーブルを Query adapter から横断して集計できる。

## 対象範囲

### やること

- 今日のタスク完了数、未完了数、完了率を取得できる。
- 今日の娯楽時間と目標との差分を取得できる。
- 日次サマリーを取得できる。
- 過去 7 日分のサマリーを取得できる。

### やらないこと

- 専用の統計保存テーブルは作らない。
- 高度なグラフ分析や月次集計は扱わない。
- タスクや娯楽時間の CRUD 自体は別 Issue で扱う。

## 実装方針

- 対象モジュール: `statistics`, `api`
- 想定ブランチ: `feature/statistics-summary`
- 主な変更ファイル/ディレクトリ:
  - `internal/statistics/`
  - `cmd/api/`
- 依存する Issue: `04-task-management.md`, `05-screentime-recording.md`
- 後続 Issue: `02-character-chat.md`

集計は Query interface と adapter で扱い、service が他モジュールの内部実装に依存しないようにする。

### 作成するファイル

```
internal/statistics/
├── model.go          ← DailyStats / WeeklyStats 定義
├── service.go        ← StatisticsQuery interface + StatisticsService
├── handler.go        ← HTTP handler（3エンドポイント）
└── adapter/
    └── postgres.go   ← PostgreSQL 横断クエリ実装
```

### model.go

```go
type DailyStats struct {
    Date                       time.Time
    CompletedTaskCount         int
    TodoTaskCount              int
    TotalTaskCount             int
    TaskCompletionRate         int  // 0-100（%）
    EntertainmentMinutes       int
    TargetEntertainmentMinutes int
    EntertainmentDiffMinutes   int
}

type WeeklyStats struct {
    From time.Time
    To   time.Time
    Days []DailyStats
}
```

### service.go

```go
type StatisticsQuery interface {
    GetDailyStats(ctx context.Context, userID string, date time.Time) (DailyStats, error)
}

type StatisticsService struct { query StatisticsQuery }

func (s *StatisticsService) GetToday(ctx, userID) (DailyStats, error)
func (s *StatisticsService) GetDaily(ctx, userID, date) (DailyStats, error)
func (s *StatisticsService) GetWeekly(ctx, userID, endDate) (WeeklyStats, error)
// GetWeekly は endDate から 7 日分を GetDailyStats でループ集計
```

### adapter/postgres.go

- `tasks` テーブルを `user_id` + `created_at::date = $date` で集計する。
  - `status = 'done'` の件数 → `CompletedTaskCount`
  - `status = 'todo'` の件数 → `TodoTaskCount`
  - 完了率 = done / total × 100（total = 0 なら 0）
- `screentime_records` テーブルを `user_id` + `date = $date` で取得する。
  - レコードなし → `EntertainmentMinutes = 0`、`TargetEntertainmentMinutes = 0`
- `internal/task` / `internal/screentime` の実装型は import しない。テーブル名だけを知る。

### handler.go

| メソッド | パス | レスポンス型 |
| --- | --- | --- |
| `GET` | `/api/v1/users/{userId}/stats/today` | `tasks` / `entertainment` のネスト構造 + `summaryText` |
| `GET` | `/api/v1/users/{userId}/stats/daily/{date}` | フラット構造 + `summaryText` |
| `GET` | `/api/v1/users/{userId}/stats/weekly` | `from`, `to`, `days[]`（`endDate` クエリパラメータ、省略時は今日） |

レスポンスはすべて `{"success": true, "data": {...}, "error": null}` 形式に統一する。  
`summaryText` の生成方法は未確定（LLM 生成 or テンプレート）。本 Issue では静的テンプレートで仮実装し、チャット Issue で差し替えを検討する。

### cmd/api/ への追加

- statistics handler のルート登録を行う。
- `PostgresStatisticsQuery` を DI で注入する。

## 受け入れ条件

- [ ] 今日の統計を取得できる。
- [ ] 指定日の統計を取得できる。
- [ ] 過去 7 日分の週次統計を取得できる。
- [ ] タスク未登録や娯楽時間未登録の日でもレスポンス形式が安定している。
- [ ] 統計専用テーブルを追加せず、既存記録から計算している。

## API / DB 変更

### API

- `GET /api/v1/users/{userId}/stats/today`
- `GET /api/v1/users/{userId}/stats/daily/{date}`
- `GET /api/v1/users/{userId}/stats/weekly?endDate=YYYY-MM-DD`（`endDate` 省略時は今日）

### DB

- 新規の統計保存テーブルは追加しない。
- `tasks` と `screentime_records` を参照して都度計算する。

## テスト・動作確認

- [ ] 今日の統計取得の正常系を確認する。
- [ ] 指定日統計取得の正常系を確認する。
- [ ] 過去 7 日分統計取得の正常系を確認する。
- [ ] データが存在しない日のレスポンスを確認する。
- [ ] 完了率と娯楽時間差分の計算を確認する。

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/api-guidelines.md`
- `docs/database-guidelines.md`
- `docs/architecture-guidelines.md`
- `docs/detail/api_design_mvp.md`

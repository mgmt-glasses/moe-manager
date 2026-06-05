# チャットログを確認できる

## 目的

ユーザが過去のチャット履歴を日付別に確認できるようにする。
ユーザ発言、キャラ返答、発言日時、ボイス再生対象を保存・取得できる状態にする。

## 背景

MVP では、秘書キャラとの対面ログを振り返れることが体験の継続性につながる。
チャット生成、ボイス生成と連携するが、ログ取得 API は履歴参照に責務を絞る。

## 対象範囲

### やること

- チャット送信時にユーザ発言とキャラ返答を保存できる。
- ユーザごとのチャットログ一覧を取得できる。
- 日付単位でチャットログを絞り込める。
- ログが存在する日付一覧を取得できる。
- ボイス再生対象テキストまたは音声参照をログに含められる。

### やらないこと

- ログの全文検索は扱わない。
- ログの編集機能は扱わない。
- 長期記憶や会話要約は扱わない。

## 実装方針

- 対象モジュール: `chat`, `api`
- 想定ブランチ: `feature/chat-logs`
- 主な変更ファイル/ディレクトリ:
  - `internal/chat/`
  - `internal/chat/adapter/`
  - `cmd/api/`
  - `prisma/migrations/`
- 依存する Issue: `02-character-chat.md`
- 後続 Issue: なし

チャットログの保存・取得責務は、チャット機能と同じ `internal/chat` に寄せる。
API 層はユーザ ID、選択キャラ、発言、返答、音声参照を渡す入口に徹する。

想定する主な interface:

```go
type ChatLogRepository interface {
	Save(ctx context.Context, log ChatLog) error
	List(ctx context.Context, query ChatLogQuery) ([]ChatLog, error)
	ListDates(ctx context.Context, userID string) ([]ChatLogDate, error)
	AttachVoice(ctx context.Context, chatLogID string, voiceFileID string) error
}
```

日付による絞り込みは JST の暦日として扱い、DB には `timestamptz` で保存する。
ユーザ発言は LLM 呼び出し前、キャラ返答は生成成功後に保存する。
ボイス生成に成功した場合は、保存済みキャラ返答へ `voice_file_id` を関連付ける。

## 受け入れ条件

- [ ] チャット送信時にユーザ発言とキャラ返答が保存される。
- [ ] ユーザごとのチャットログを取得できる。
- [ ] 日付でチャットログを絞り込める。
- [ ] ログが存在する日付一覧を取得できる。
- [ ] 他ユーザのログを取得できない。
- [ ] ボイス付き返答の場合、音声参照をログから確認できる。

## API / DB 変更

### API

- `GET /api/v1/users/{userId}/chat/messages`
- `GET /api/v1/users/{userId}/chat-logs`
- `GET /api/v1/users/{userId}/chat-logs/dates`

### DB

- チャットログ保存用テーブルを扱う。
- ユーザ ID、キャラ ID、発言者、メッセージ、発言日時、音声参照を保存する。
- 日付別取得に必要なインデックスを検討する。

## テスト・動作確認

- [ ] チャット送信後にログが保存されることを確認する。
- [ ] ログ一覧取得の正常系を確認する。
- [ ] 日付指定取得の正常系を確認する。
- [ ] ログ日付一覧取得の正常系を確認する。
- [ ] 他ユーザのログが混ざらないことを確認する。
- [ ] Repository はテスト用 PostgreSQL または transaction rollback を用いて検証する。

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/api-guidelines.md`
- `docs/database-guidelines.md`
- `docs/detail/api_design_mvp.md`

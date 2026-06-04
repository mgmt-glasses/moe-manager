# 選んだ秘書キャラとチャットできる

## 目的

ユーザが選択中の秘書キャラにメッセージを送り、キャラ設定に沿った返答を受け取れるようにする。
返答生成では、ユーザのタスク状況や娯楽時間を必要に応じて参照できる設計にする。

## 背景

MVP の中心体験は、ユーザの行動に対して秘書キャラが会話で反応することにある。
チャット生成と他ドメイン情報の統合は Go の `internal/chat` の責務とし、`cmd/api` は依存性注入に徹する。

## 対象範囲

### やること

- ユーザ発言を送信し、秘書キャラの返答を取得できる。
- 選択中キャラの MBTI、性格、口調をプロンプトに反映する。
- タスク状況と娯楽時間をプロンプト用コンテキストとして渡せる。
- 返答テキストをチャットログ保存に渡せる構造にする。

### やらないこと

- ボイスファイル生成はこの Issue では扱わない。
- チャットログ一覧表示 API は別 Issue で扱う。
- 複雑な長期記憶や会話要約は扱わない。

## 実装方針

- 対象モジュール: `chat`, `api`
- 想定ブランチ: `feature/chat-character-response`
- 主な変更ファイル/ディレクトリ:
  - `internal/chat/`
  - `cmd/api/`
- 依存する Issue: `01-character-selection.md`, `04-task-management.md`, `05-screentime-recording.md`
- 後続 Issue: `03-character-voice.md`, `07-chat-logs.md`

`internal/chat` は user、character、task、screentime の情報を取得する interface と、チャット生成に必要なコンテキスト型を自身で定義する。
`cmd/api` は各 interface の実装を組み立てるだけとし、チャット処理の手順を持たない。

想定する主な interface:

```go
type LLMClient interface {
	GenerateReply(ctx context.Context, prompt Prompt) (string, error)
}

type ContextLoader interface {
	Load(ctx context.Context, userID string) (ChatContext, error)
}
```

LLM プロバイダ固有の API キーやモデル選択をリクエストヘッダから直接受け取らず、サーバー設定と adapter に閉じる。

## 受け入れ条件

- [ ] ユーザがメッセージを送信できる。
- [ ] 選択中キャラの設定に基づく返答テキストが返る。
- [ ] 未選択キャラなど必要情報が不足している場合は適切なエラーになる。
- [ ] 返答は共通レスポンス形式で返る。
- [ ] チャット生成ロジックが API 層ではなく `internal/chat` に閉じている。

## API / DB 変更

### API

- `POST /api/v1/users/{userId}/chat/messages`

### DB

- この Issue 単体では新規 DB 永続化を必須にしない。
- チャットログ保存を同時に行う場合は `07-chat-logs.md` と整合させる。

## テスト・動作確認

- [ ] チャット送信の正常系を確認する。
- [ ] 選択中キャラが返答生成に反映されることを確認する。
- [ ] タスク状況または娯楽時間がコンテキストとして渡ることを確認する。
- [ ] 必須情報不足時の異常系を確認する。
- [ ] LLM は fake adapter に差し替えて `go test ./...` で確認できる。

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/api-guidelines.md`
- `docs/architecture-guidelines.md`
- `docs/detail/api_design_mvp.md`

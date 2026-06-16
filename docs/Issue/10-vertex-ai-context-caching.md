# Vertex AI Context Caching で LLM トークンコストを削減する

## 目的

チャット送信のたびにキャラクター設定の system instruction を毎回 Vertex AI へ送信しているトークンコストを、Context Caching によって削減する。

## 背景

`POST /api/v1/users/{userId}/chat/messages` は毎リクエストで以下をモデルに渡す。

- system instruction（キャラ名・MBTI・性格・口調・ルール）
- ユーザのタスク状況・娯楽時間
- 会話履歴（直近 N ターン）
- ユーザ発言

このうち system instruction はキャラクター設定に依存し、同一ユーザが同一キャラクターとチャットしている間は内容が変わらない。
Vertex AI の Context Caching を使用すると、system instruction をサーバー側にキャッシュし、後続リクエストではキャッシュ参照 ID だけを送ればよくなる。
キャッシュを使わない場合と比較して、system instruction 分の入力トークンコストを削減できる。

## 対象範囲

### やること

- `internal/chat/adapter/` に Vertex AI LLMClient adapter を実装する。
- adapter 内で system instruction の Context Cache を管理する（作成・参照・TTL 更新）。
- キャッシュが有効な間は system instruction をリクエストボディに含めず、キャッシュ参照 ID を使う。
- キャッシュが期限切れまたは存在しない場合は自動的に再作成する。
- キャッシュの保存先は adapter 内のインメモリ map とし、プロセス再起動で再作成する設計にする。

### やらないこと

- キャッシュ ID の永続化（PostgreSQL や Redis への保存）は行わない。
- キャッシュ TTL の外部設定化はこの Issue では行わない。
- `PersonaRepository` の Firestore adapter は実装しない（PostgreSQL の `characters` テーブルを使用する）。
- 会話履歴のキャッシュは行わない（内容が毎リクエスト変わるため）。

## 実装方針

- 対象モジュール: `chat`, `api`
- 想定ブランチ: `feature/chat-vertex-ai-caching`
- 主な変更ファイル/ディレクトリ:
  - `internal/chat/adapter/vertexai_llm.go`（新規）
  - `cmd/api/`（依存性注入の更新）
- 依存する Issue: `02-character-chat.md`
- 後続 Issue: なし

### キャッシュ管理の設計

キャッシュ管理は `LLMClient` adapter の内部に閉じる。`Service` や `Prompt` 型は変更しない。

```
Service.SendMessage
  └─ buildPrompt → Prompt{System: "...", User: "..."}
       └─ LLMClient.GenerateReply(ctx, prompt)
            └─ adapter 内部:
                 hash(prompt.System) でキャッシュ検索
                 キャッシュなし → Vertex AI にキャッシュ作成 → cacheID 保存
                 キャッシュあり → cacheID を使ってリクエスト送信
```

インメモリキャッシュのデータ構造:

```go
type cacheEntry struct {
    cacheID   string
    expiresAt time.Time
}
// key: hash of system instruction
var cache map[string]cacheEntry
```

### Vertex AI Context Caching の制約

- キャッシュ対象の最小トークン数が Vertex AI 側で設定されている（モデルによって異なる）。
- TTL のデフォルトは 1 時間（設定可能）。
- キャッシュ対象は system instruction のみとし、会話履歴は毎回送信する。

### PostgreSQL からのキャラ設定取得

`PersonaRepository` は PostgreSQL の `characters` テーブルから `personality_desc`、`speech_style`、`system_prompt_fragment` を取得する adapter として実装する。
GCP Firestore は使用しない。

```go
// internal/chat/adapter/postgres_persona.go
func (r *PostgresPersonaRepository) Get(ctx context.Context, characterID string) (chat.Persona, error) {
    // characters テーブルから personality_desc, speech_style, system_prompt_fragment を取得
    // Persona.SystemInstruction として組み立てて返す
}
```

## 受け入れ条件

- [ ] Vertex AI LLMClient adapter が実装されている。
- [ ] 同一の system instruction を持つリクエストで、2回目以降はキャッシュ参照 ID を使用している。
- [ ] キャッシュが期限切れの場合、自動的に再作成して処理を継続する。
- [ ] `PersonaRepository` の PostgreSQL adapter が実装されている。
- [ ] `cmd/api` で依存性注入が完成し、チャット API がエンドツーエンドで動作する。
- [ ] LLMClient は fake adapter に差し替えて `go test ./...` が通る。

## API / DB 変更

### API

- 変更なし。

### DB

- 変更なし。キャッシュ ID はインメモリ管理のため DB カラム追加は不要。

## テスト・動作確認

- [ ] fake LLMClient を使ったユニットテストが通る。
- [ ] Vertex AI に実際に接続し、キャッシュが作成されることを確認する。
- [ ] 2回目のチャット送信でキャッシュ参照 ID が使われることをログで確認する。
- [ ] プロセス再起動後にキャッシュが自動再作成されることを確認する。
- [ ] キャラクターが変わった場合（system instruction が変わった場合）に別キャッシュが作成されることを確認する。

## 参照ドキュメント

- `docs/Issue/02-character-chat.md`
- `docs/architecture-guidelines.md`
- `docs/Issue/09-postgresql-schema-design.md`
- [Vertex AI Context Caching ドキュメント](https://cloud.google.com/vertex-ai/generative-ai/docs/context-cache/context-cache-overview)

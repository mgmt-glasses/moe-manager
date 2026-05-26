# 選んだ秘書キャラとチャットできる

## 目的

ユーザが選択中の秘書キャラにメッセージを送り、キャラ設定に沿った返答を受け取れるようにする。
返答生成では、ユーザのタスク状況や娯楽時間を必要に応じて参照できる設計にする。

## 背景

MVP の中心体験は、ユーザの行動に対して秘書キャラが会話で反応することにある。
チャット生成自体は `mkh_voice` の責務とし、他モジュール情報の取得・統合は gateway 層で扱う。

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

- 対象モジュール: `mkh_voice`, `gateway`
- 想定ブランチ: `feature/voice-character-chat`
- 主な変更ファイル/ディレクトリ:
  - `packages/voice-library/`
  - `apps/gateway/gateway/main.py`
- 依存する Issue: `01-character-selection.md`, `04-task-management.md`, `05-screentime-recording.md`
- 後続 Issue: `03-character-voice.md`, `07-chat-logs.md`

gateway で `moe_user`, `moe_character`, `moe_task`, `moe_screentime` から必要情報を集約し、`mkh_voice` にプリミティブなコンテキストとして渡す。

## 受け入れ条件

- [ ] ユーザがメッセージを送信できる。
- [ ] 選択中キャラの設定に基づく返答テキストが返る。
- [ ] 未選択キャラなど必要情報が不足している場合は適切なエラーになる。
- [ ] 返答は共通レスポンス形式で返る。
- [ ] チャット生成ロジックが gateway ではなく `mkh_voice` 側に閉じている。

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

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/api-guidelines.md`
- `docs/architecture-guidelines.md`
- `docs/detail/api_design_mvp.md`

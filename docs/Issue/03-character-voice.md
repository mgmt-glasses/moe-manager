# キャラの返答にボイスが付く

## 目的

秘書キャラのチャット返答に対して、キャラ別のボイスを生成または取得できるようにする。
フロントエンドは音声生成の詳細を意識せず、返答テキストと音声 URL を扱える状態にする。

## 背景

MVP では、秘書キャラがテキストだけでなく声でも反応することが重要な体験になる。
キャラごとのボイス設定は `voice_preset_id` を正とし、API では `voiceId` として返す。

## 対象範囲

### やること

- チャット返答テキストから音声を生成できる。
- 選択中キャラの `voice_preset_id` を使ってボイスを選択する。
- 生成済み音声を取得できる URL または ID を返す。
- API 利用者に TTS 実装詳細を露出しない。

### やらないこと

- ボイスサンプル管理の詳細機能は扱わない。
- 音声ストレージの高度なライフサイクル管理は扱わない。
- チャット返答の生成ロジック自体は別 Issue で扱う。

## 実装方針

- 対象モジュール: `mkh_voice`, `gateway`
- 想定ブランチ: `feature/voice-response-audio`
- 主な変更ファイル/ディレクトリ:
  - `packages/voice-library/`
  - `apps/gateway/gateway/main.py`
- 依存する Issue: `01-character-selection.md`, `02-character-chat.md`
- 後続 Issue: `07-chat-logs.md`

ボイス生成は `mkh_voice` の Port / Adapter に閉じ、gateway は選択中キャラと返答テキストを渡すだけにする。

## 受け入れ条件

- [ ] キャラ返答テキストに対応する音声を生成できる。
- [ ] 選択中キャラのボイス設定が使われる。
- [ ] API レスポンスに音声参照情報が含まれる。
- [ ] 音声ファイル取得 API で生成済み音声を取得できる。
- [ ] TTS 失敗時にチャット返答全体をどう扱うかが明確になっている。

## API / DB 変更

### API

- `POST /api/v1/users/{userId}/voices`
- `GET /api/v1/voice-files/{voiceFileId}`
- チャット API のレスポンスに音声参照を含める可能性がある。

### DB

- 必要に応じて音声ファイル ID、生成元テキスト、キャラ、作成日時を保存する。
- 保存する場合はチャットログとの関連付けを `07-chat-logs.md` と整合させる。

## テスト・動作確認

- [ ] 返答テキストから音声生成できることを確認する。
- [ ] キャラ別の `voice_preset_id` が使われることを確認する。
- [ ] 音声取得 API の正常系を確認する。
- [ ] TTS 失敗時のエラー形式を確認する。

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/api-guidelines.md`
- `docs/implementation-guide.md`
- `docs/detail/api_design_mvp.md`

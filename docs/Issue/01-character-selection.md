# ユーザが MBTI 別の秘書キャラを選べる

## 目的

ユーザが初期設定または設定変更で、MBTI 別の秘書キャラを選択できるようにする。
選択したキャラはユーザ設定として保存し、チャットやボイス生成で参照できる状態にする。

## 背景

MVP の中核体験である「社長の行動に秘書キャラが反応してくれる」を成立させるため、ユーザごとに使用キャラを決められる必要がある。
キャラ情報は `moe_character`、選択状態は `moe_user` の責務として扱う。

## 対象範囲

### やること

- MBTI 別の秘書キャラ一覧を取得できる。
- キャラ詳細を取得できる。
- ユーザの選択中キャラを保存・更新できる。
- API レスポンスではキャラの MBTI、表示名、説明、口調、ボイス設定参照を返す。

### やらないこと

- 全 16 MBTI タイプの完全実装は必須にしない。
- キャラ画像や高度なプロフィール管理はこの Issue では扱わない。
- チャット返答生成とボイス生成は別 Issue で扱う。

## 実装方針

- 対象モジュール: `moe_character`, `moe_user`, `gateway`
- 想定ブランチ: `feature/character-selection`
- 主な変更ファイル/ディレクトリ:
  - `packages/character/`
  - `packages/user/`
  - `apps/gateway/gateway/main.py`
- 依存する Issue: なし
- 後続 Issue: `02-character-chat.md`, `03-character-voice.md`

複数モジュールにまたがるため、PR を分ける場合はキャラ一覧 API とユーザ選択保存 API を別 Issue に分割する。

## 受け入れ条件

- [ ] ユーザが利用可能な秘書キャラ一覧を取得できる。
- [ ] キャラ詳細に MBTI、説明、口調、`voiceId` が含まれる。
- [ ] ユーザが選択中キャラを更新できる。
- [ ] ユーザ情報取得時に選択中キャラを判別できる。
- [ ] 存在しないキャラ ID を指定した場合は適切なエラーになる。

## API / DB 変更

### API

- `GET /api/v1/characters`
- `GET /api/v1/characters/{characterId}`
- `PATCH /api/v1/users/{userId}/selected-character`

### DB

- キャラマスタを扱う。
- ユーザ設定に選択中キャラ ID を保存する。
- ボイス設定は domain/DB では `voice_preset_id`、API では `voiceId` として扱う。

## テスト・動作確認

- [ ] キャラ一覧取得の正常系を確認する。
- [ ] キャラ詳細取得の正常系と存在しない ID の異常系を確認する。
- [ ] 選択中キャラ更新の正常系を確認する。
- [ ] 更新後のユーザ情報に選択キャラが反映されることを確認する。

## 参照ドキュメント

- `docs/mvp-summary.md`
- `docs/api-guidelines.md`
- `docs/implementation-guide.md`
- `docs/detail/api_design_mvp.md`

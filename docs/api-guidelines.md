# API 方針

このドキュメントは、既存の `detail/api_design_mvp.md` をもとに API 設計の共通方針を整理したものです。

## 基本方針

- フロントエンドから扱いやすい REST API にする。
- 画面単位ではなく、ドメイン単位で API を分ける。
- AI 生成やボイス生成の詳細はバックエンド側に隠す。
- フロントエンドは、チャット送信と音声 URL 取得を中心に扱える設計にする。
- 統計はリアルタイム計算を基本とし、日次集計テーブルへの保存は将来対応とする。
- ユーザー別 API は Firebase ID token による認証を必須にし、トークンの `uid` とパス上の `userId` を照合する。

## Base URL

```txt
/api/v1
```

## 共通レスポンス

認証が必要な API は以下のヘッダを必須にします。

```txt
Authorization: Bearer <Firebase ID token>
```

成功時:

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

エラー時:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "入力内容が正しくありません"
  }
}
```

## 日付と日時

日付:

```txt
YYYY-MM-DD
```

日時:

```txt
ISO 8601
```

例:

```txt
2026-05-13T12:30:00+09:00
```

## API ドメイン

MVP で扱う API ドメインは以下です。

| ドメイン | 役割 |
| --- | --- |
| Auth / User | 初期ユーザ作成、ユーザ情報取得、設定更新 |
| Characters | キャラ一覧、キャラ詳細、使用キャラ変更 |
| User Settings | 呼び名、目標娯楽時間、選択キャラなど |
| Tasks | タスク登録、一覧、完了、削除 |
| Entertainment Time | 娯楽時間の記録、目標との差分 |
| Stats | 今日、週次、日次サマリー |
| Chat | ユーザ発言送信、AI 応答生成 |
| Voice | 返答テキストの音声生成、音声 URL |
| Chat Logs | 会話履歴の保存と取得 |

## 優先実装順

実装済み API の詳細なリクエスト/レスポンス例は `api-reference.md` を参照します。
未実装 API の検討時は `detail/api_design_mvp.md` を参考にします。
実装順は以下を基本とします。

| 順序 | 対象 | API |
| --- | --- | --- |
| 1 | 基盤 | `POST /users`, `GET /users/{userId}`, `PATCH /users/{userId}`, `GET /characters`, `GET /characters/{characterId}`, `PATCH /users/{userId}/selected-character` |
| 2 | タスク | `POST /users/{userId}/tasks`, `GET /users/{userId}/tasks`, `PATCH /users/{userId}/tasks/{taskId}/complete`, `PATCH /users/{userId}/tasks/{taskId}/reopen`, `DELETE /users/{userId}/tasks/{taskId}` |
| 3 | 娯楽時間 | `POST /users/{userId}/screentime/{date}`, `GET /users/{userId}/screentime/{date}`, `GET /users/{userId}/screentime`, `POST /users/{userId}/screentime/analyze` |
| 4 | 統計 | `GET /users/{userId}/stats/today`, `GET /users/{userId}/stats/daily/{date}`, `GET /users/{userId}/stats/weekly` |
| 5 | チャット | `POST /users/{userId}/chat/messages`, `GET /users/{userId}/chat/messages`, `GET /users/{userId}/chat-logs`, `GET /users/{userId}/chat-logs/dates` |
| 6 | ボイス | `POST /users/{userId}/voices`, `GET /users/{userId}/voice-files/{voiceFileId}` |

パスは Base URL `/api/v1` 配下として扱います。

## AI 返答の制約

チャット API では、AI に以下の制約を渡します。

- ユーザを社長として扱う。
- キャラの MBTI、性格、口調を守る。
- タスク状況と娯楽時間に必要に応じて触れる。
- 必要以上に責めない。
- 返答は短く自然な会話にする。
- 医療、法律、金融などの専門助言は避ける。
- タスク管理アプリの秘書として振る舞う。

## 設計時の注意

- API は domain の use case を呼び出す入口に徹する。
- Go の HTTP handler にビジネスロジックを置かない。
- レスポンスのキー名はフロントエンド都合で一貫させる。
- DB のカラム名と API のフィールド名が異なる場合は、変換責務を adapter 側に閉じる。
- チャットやボイス生成の実装詳細を API 利用者に露出させない。
- ボイス設定は domain/DB では `voice_preset_id`、API では `voiceId` として扱う。

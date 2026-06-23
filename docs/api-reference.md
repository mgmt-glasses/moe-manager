# API リファレンス

フロントエンド向け API リファレンスです。現在実装済みのエンドポイントをすべて記載しています。

## 目次

- [基本情報](#基本情報)
- [共通レスポンス形式](#共通レスポンス形式)
- [エラーコード一覧](#エラーコード一覧)
- [ユーザー](#ユーザー)
- [キャラクター](#キャラクター)
- [タスク](#タスク)
- [統計](#統計)
- [チャット](#チャット)
- [音声生成](#音声生成)

---

## 基本情報

| 項目 | 値 |
|------|-----|
| ベース URL（ローカル） | `http://localhost:8081` |
| Content-Type | `application/json`（音声取得エンドポイントを除く） |
| 認証 | なし（MVP フェーズ） |

### ローカル起動手順

```bash
# 環境変数を設定（.env を参照）
export DATABASE_URL=postgres://...
export VERTEX_PROJECT=your-gcp-project
export PORT=8081

go run ./cmd/api
```

---

## 共通レスポンス形式

すべての JSON エンドポイントは以下の形式を返します。

```json
{
  "success": true,
  "data": { ... },
  "error": null
}
```

エラー時：

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーの説明"
  }
}
```

---

## エラーコード一覧

| コード | HTTP ステータス | 説明 |
|--------|----------------|------|
| `INVALID_BODY` | 400 | リクエストボディのパース失敗 |
| `VALIDATION_ERROR` | 400 | 必須フィールドの不足・形式不正 |
| `INVALID_DATE` | 400 | 日付フォーマット不正（YYYY-MM-DD 形式が必要） |
| `CHARACTER_NOT_FOUND` | 400 | 指定したキャラクターが存在しない |
| `NOT_FOUND` | 404 | リソースが見つからない |
| `FORBIDDEN` | 403 | 操作権限なし |
| `NO_CHARACTER_SELECTED` | 422 | キャラクターが未選択 |
| `PERSONA_NOT_FOUND` | 422 | キャラクター設定が未登録 |
| `INTERNAL_ERROR` | 500 | サーバー内部エラー |

---

## ユーザー

### ユーザー作成

```
POST /api/v1/users
```

**リクエスト**

```json
{
  "name": "テスト社長",
  "presidentName": "田中",
  "targetEntertainmentMinutes": 60,
  "selectedCharacterId": "char_enfj_001"
}
```

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `name` | string | ✓ | ユーザー名 |
| `presidentName` | string | ✓ | 社長名（チャットで使用） |
| `targetEntertainmentMinutes` | number | ✓ | 娯楽時間の目標（分） |
| `selectedCharacterId` | string | - | 選択キャラクター ID |

**レスポンス** `201 Created`

```json
{
  "success": true,
  "data": {
    "userId": "7c1e710b-...",
    "name": "テスト社長",
    "presidentName": "田中",
    "targetEntertainmentMinutes": 60,
    "selectedCharacterId": "char_enfj_001",
    "createdAt": "2026-06-20T10:00:00+09:00",
    "updatedAt": "2026-06-20T10:00:00+09:00"
  },
  "error": null
}
```

---

### ユーザー取得

```
GET /api/v1/users/{userId}
```

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": {
    "userId": "7c1e710b-...",
    "name": "テスト社長",
    "presidentName": "田中",
    "targetEntertainmentMinutes": 60,
    "selectedCharacterId": "char_enfj_001",
    "createdAt": "2026-06-20T10:00:00+09:00",
    "updatedAt": "2026-06-20T10:00:00+09:00"
  },
  "error": null
}
```

---

### ユーザー更新

```
PATCH /api/v1/users/{userId}
```

**リクエスト**（すべてのフィールドは任意。送信したフィールドのみ更新）

```json
{
  "name": "新しい名前",
  "presidentName": "鈴木",
  "targetEntertainmentMinutes": 90,
  "selectedCharacterId": "char_intj_001"
}
```

**レスポンス** `200 OK` — ユーザー取得と同じ形式

---

### 選択キャラクター変更

```
PATCH /api/v1/users/{userId}/selected-character
```

**リクエスト**

```json
{
  "characterId": "char_enfj_001"
}
```

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": {
    "userId": "7c1e710b-...",
    "selectedCharacterId": "char_enfj_001"
  },
  "error": null
}
```

---

## キャラクター

### キャラクター一覧取得

```
GET /api/v1/characters
```

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "characterId": "char_enfj_001",
      "name": "ルナ",
      "mbti": "ENFJ",
      "description": "社交的で共感力が高いキャラクター",
      "tone": "明るく丁寧",
      "voiceId": "preset_01",
      "iconUrl": "/assets/characters/enfj/icon.png",
      "standingImageUrl": "/assets/characters/enfj/standing.png"
    }
  ],
  "error": null
}
```

---

### キャラクター詳細取得

```
GET /api/v1/characters/{characterId}
```

**レスポンス** `200 OK`（一覧に `sampleVoiceUrl` が追加）

```json
{
  "success": true,
  "data": {
    "characterId": "char_enfj_001",
    "name": "ルナ",
    "mbti": "ENFJ",
    "description": "社交的で共感力が高いキャラクター",
    "tone": "明るく丁寧",
    "voiceId": "preset_01",
    "iconUrl": "/assets/characters/enfj/icon.png",
    "standingImageUrl": "/assets/characters/enfj/standing.png",
    "sampleVoiceUrl": "/assets/characters/enfj/sample.wav"
  },
  "error": null
}
```

---

## タスク

### タスク作成

```
POST /api/v1/users/{userId}/tasks
```

**リクエスト**

```json
{
  "title": "資料作成"
}
```

**レスポンス** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "abc123-...",
    "title": "資料作成",
    "status": "todo",
    "createdAt": "2026-06-20T10:00:00+09:00",
    "completedAt": null
  },
  "error": null
}
```

`status` の値: `"todo"` | `"done"`

---

### タスク一覧取得

```
GET /api/v1/users/{userId}/tasks
```

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "abc123-...",
      "title": "資料作成",
      "status": "todo",
      "createdAt": "2026-06-20T10:00:00+09:00",
      "completedAt": null
    }
  ],
  "error": null
}
```

---

### タスク完了

```
PATCH /api/v1/users/{userId}/tasks/{taskId}/complete
```

リクエストボディ不要。

**レスポンス** `200 OK` — タスク作成と同じ形式（`status: "done"`, `completedAt` に日時が入る）

---

### タスク再オープン

```
PATCH /api/v1/users/{userId}/tasks/{taskId}/reopen
```

リクエストボディ不要。

**レスポンス** `200 OK` — タスク作成と同じ形式（`status: "todo"`, `completedAt: null`）

---

### タスク削除

```
DELETE /api/v1/users/{userId}/tasks/{taskId}
```

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": null,
  "error": null
}
```

---

## 統計

### 今日の統計取得

```
GET /api/v1/users/{userId}/stats/today
```

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": {
    "date": "2026-06-20",
    "tasks": {
      "completedCount": 3,
      "todoCount": 2,
      "totalCount": 5,
      "completionRate": 60
    },
    "entertainment": {
      "minutes": 45,
      "targetMinutes": 60,
      "diffMinutes": -15
    },
    "summaryText": "今日は5件中3件のタスクを完了しています。娯楽時間は目標より15分少なく、良好です。"
  },
  "error": null
}
```

`diffMinutes` が正の値 → 目標超過、負の値 → 目標未満（良好）

---

### 特定日の統計取得

```
GET /api/v1/users/{userId}/stats/daily/{date}
```

`{date}` は `YYYY-MM-DD` 形式。

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": {
    "date": "2026-06-20",
    "completedTaskCount": 3,
    "todoTaskCount": 2,
    "totalTaskCount": 5,
    "taskCompletionRate": 60,
    "entertainmentMinutes": 45,
    "targetEntertainmentMinutes": 60,
    "entertainmentDiffMinutes": -15,
    "summaryText": "..."
  },
  "error": null
}
```

---

### 週間統計取得

```
GET /api/v1/users/{userId}/stats/weekly?endDate=2026-06-20
```

| クエリパラメータ | 型 | 必須 | 説明 |
|----------------|-----|------|------|
| `endDate` | string | - | 集計終了日（YYYY-MM-DD）。省略時は今日 |

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": {
    "from": "2026-06-14",
    "to": "2026-06-20",
    "days": [
      {
        "date": "2026-06-14",
        "completedTaskCount": 2,
        "todoTaskCount": 1,
        "taskCompletionRate": 67,
        "entertainmentMinutes": 30,
        "targetEntertainmentMinutes": 60,
        "entertainmentDiffMinutes": -30
      }
    ]
  },
  "error": null
}
```

---

## チャット

キャラクターとのチャット（Vertex AI Gemini によるレスポンス生成）。

### メッセージ送信

```
POST /api/v1/users/{userId}/chat/messages
```

**前提条件**: ユーザーがキャラクターを選択済みであること（`selectedCharacterId` が設定されていること）

**リクエスト**

```json
{
  "message": "今日のタスク進捗を教えて"
}
```

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": {
    "userMessage": {
      "role": "user",
      "message": "今日のタスク進捗を教えて",
      "createdAt": "2026-06-20T10:00:00+09:00"
    },
    "assistantMessage": {
      "role": "assistant",
      "message": "今日は5件中3件完了しているね！残り2件、一緒に頑張ろう！",
      "createdAt": "2026-06-20T10:00:01+09:00"
    },
    "context": {
      "todayTaskCompletedCount": 3,
      "todayTaskTotalCount": 5,
      "todayEntertainmentMinutes": 45,
      "targetEntertainmentMinutes": 60
    }
  },
  "error": null
}
```

**エラー例**

| ケース | コード | HTTP |
|--------|--------|------|
| キャラクター未選択 | `NO_CHARACTER_SELECTED` | 422 |
| キャラクター設定未登録 | `PERSONA_NOT_FOUND` | 422 |
| ユーザー不在 | `NOT_FOUND` | 404 |

---

## 音声生成

選択中のキャラクターの声でテキストを読み上げる WAV ファイルを生成します。

> **前提**: TTS サービス（Python voice-library）が `TTS_SERVICE_URL`（デフォルト: `http://localhost:8001`）で起動していること。

### 音声生成

```
POST /api/v1/users/{userId}/voices
```

**リクエスト**

```json
{
  "characterId": "char_enfj_001",
  "voicePresetId": "preset_01",
  "text": "今日もお疲れ様です！"
}
```

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `characterId` | string | ✓ | ユーザーの選択キャラクターと一致する必要あり |
| `voicePresetId` | string | ✓ | キャラクターの `voiceId` を使用 |
| `text` | string | ✓ | 読み上げるテキスト |

**レスポンス** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "3d0e59eb-...",
    "characterId": "char_enfj_001",
    "sourceText": "今日もお疲れ様です！",
    "createdAt": "2026-06-20T10:00:00+09:00"
  },
  "error": null
}
```

**エラー例**

| ケース | コード | HTTP |
|--------|--------|------|
| 他ユーザーのキャラクター指定 | `FORBIDDEN` | 403 |
| キャラクター未選択 | `NO_CHARACTER_SELECTED` | 422 |
| TTS サービス未起動 | `INTERNAL_ERROR` | 500 |

---

### 音声ファイル取得

```
GET /api/v1/users/{userId}/voice-files/{voiceFileId}
```

> ⚠️ このエンドポイントは JSON ではなく **WAV バイナリ** を返します。

**レスポンスヘッダー**

```
Content-Type: audio/wav
```

**フロントエンドでの使用例**

```javascript
const res = await fetch(`/api/v1/users/${userId}/voice-files/${voiceFileId}`);
const blob = await res.blob();
const url = URL.createObjectURL(blob);
const audio = new Audio(url);
audio.play();
```

**エラー時のみ JSON を返します**

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "NOT_FOUND",
    "message": "音声ファイルが見つかりません"
  }
}
```

---

## 娯楽時間管理 (ScreenTime)

### 画像解析 (Gemini API)

スクリーンタイムの画像を解析し、利用時間を自動抽出します。

```
POST /api/v1/users/{userId}/screentime/analyze
```

**リクエスト**

```json
{
  "image": "<Base64エンコードされた画像データ>",
  "mime_type": "image/jpeg"
}
```

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": {
    "minutes": 120
  },
  "error": null
}
```

---

### 娯楽時間の登録・更新

```
POST /api/v1/users/{userId}/screentime/{date}
```

`{date}` は `YYYY-MM-DD` 形式。

**リクエスト**

```json
{
  "minutes": 120,
  "target_minutes": 60
}
```

**レスポンス** `200 OK`

```json
{
  "success": true,
  "data": {
    "record_id": "abc123-...",
    "user_id": "7c1e710b-...",
    "date": "2026-06-20",
    "minutes": 120,
    "target_minutes": 60,
    "diff_minutes": 60
  },
  "error": null
}
```

---

### 娯楽時間の取得

```
GET /api/v1/users/{userId}/screentime/{date}
```

**レスポンス** `200 OK`（登録・更新と同じ形式）

---

### 娯楽時間の一覧取得

```
GET /api/v1/users/{userId}/screentime?from=2026-06-01&to=2026-06-20
```

| クエリパラメータ | 型 | 必須 | 説明 |
|----------------|-----|------|------|
| `from` | string | - | 取得開始日（YYYY-MM-DD）。省略時は7日前 |
| `to` | string | - | 取得終了日（YYYY-MM-DD）。省略時は今日 |

**レスポンス** `200 OK`（登録・更新の配列形式）

---

## 実装予定

以下は現在開発中または未実装のエンドポイントです。

| 機能 | 状態 |
|------|------|
| チャットログ保存 | 未実装（Issue 07） |
| 音声ファイル一覧取得 | 未実装 |

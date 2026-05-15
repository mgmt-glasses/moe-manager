# API設計：二次元秘書AIタスク・娯楽管理アプリ MVP

## 1. API設計方針

MVPでは、以下の方針でAPIを設計する。

- フロントエンドから扱いやすいREST APIにする
- 画面単位ではなく、ドメイン単位でAPIを分ける
- AI生成やボイス生成の詳細はバックエンド側に隠す
- フロントエンドは「チャットを送る」「音声URLを受け取る」だけでよい設計にする
- 統計はリアルタイム計算を基本とし、必要に応じて日次集計テーブルに保存する
- MVPでは認証は簡易実装でもよいが、API上は userId を前提にする

---

# 2. APIドメイン一覧

MVPで必要なAPIドメインは以下。

- Auth / User
- Characters
- User Settings
- Tasks
- Entertainment Time
- Stats
- Chat
- Voice
- Chat Logs

---

# 3. 共通仕様

## 3.1 Base URL

```txt
/api/v1
```

## 3.2 共通レスポンス形式

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

エラー時：

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

## 3.3 日付形式

```txt
YYYY-MM-DD
```

例：

```txt
2026-05-13
```

## 3.4 日時形式

```txt
ISO 8601
```

例：

```txt
2026-05-13T12:30:00+09:00
```

---

# 4. User API

## 4.1 ユーザ作成

初回起動時にユーザを作成する。

```http
POST /api/v1/users
```

### Request

```json
{
  "name": "太郎",
  "presidentName": "社長",
  "targetEntertainmentMinutes": 120,
  "selectedCharacterId": "char_istj_001"
}
```

### Response

```json
{
  "success": true,
  "data": {
    "userId": "user_001",
    "name": "太郎",
    "presidentName": "社長",
    "targetEntertainmentMinutes": 120,
    "selectedCharacterId": "char_istj_001",
    "createdAt": "2026-05-13T12:00:00+09:00"
  },
  "error": null
}
```

---

## 4.2 ユーザ情報取得

```http
GET /api/v1/users/{userId}
```

### Response

```json
{
  "success": true,
  "data": {
    "userId": "user_001",
    "name": "太郎",
    "presidentName": "社長",
    "targetEntertainmentMinutes": 120,
    "selectedCharacterId": "char_istj_001",
    "createdAt": "2026-05-13T12:00:00+09:00"
  },
  "error": null
}
```

---

## 4.3 ユーザ設定更新

```http
PATCH /api/v1/users/{userId}
```

### Request

```json
{
  "name": "太郎",
  "presidentName": "社長",
  "targetEntertainmentMinutes": 90,
  "selectedCharacterId": "char_enfj_001"
}
```

### Response

```json
{
  "success": true,
  "data": {
    "userId": "user_001",
    "name": "太郎",
    "presidentName": "社長",
    "targetEntertainmentMinutes": 90,
    "selectedCharacterId": "char_enfj_001",
    "updatedAt": "2026-05-13T13:00:00+09:00"
  },
  "error": null
}
```

---

# 5. Character API

## 5.1 キャラ一覧取得

キャラ選択画面で使用する。

```http
GET /api/v1/characters
```

### Response

```json
{
  "success": true,
  "data": [
    {
      "characterId": "char_istj_001",
      "name": "白石 澪",
      "mbti": "ISTJ",
      "description": "真面目で几帳面な管理型秘書",
      "tone": "丁寧で落ち着いた口調",
      "voiceId": "voice_istj_001",
      "iconUrl": "/assets/characters/istj/icon.png",
      "standingImageUrl": "/assets/characters/istj/standing.png"
    }
  ],
  "error": null
}
```

---

## 5.2 キャラ詳細取得

```http
GET /api/v1/characters/{characterId}
```

### Response

```json
{
  "success": true,
  "data": {
    "characterId": "char_istj_001",
    "name": "白石 澪",
    "mbti": "ISTJ",
    "description": "真面目で几帳面な管理型秘書",
    "tone": "丁寧で落ち着いた口調",
    "voiceId": "voice_istj_001",
    "iconUrl": "/assets/characters/istj/icon.png",
    "standingImageUrl": "/assets/characters/istj/standing.png",
    "sampleVoiceUrl": "/assets/voices/istj/sample.wav"
  },
  "error": null
}
```

---

## 5.3 使用キャラ変更

```http
PATCH /api/v1/users/{userId}/selected-character
```

### Request

```json
{
  "characterId": "char_enfj_001"
}
```

### Response

```json
{
  "success": true,
  "data": {
    "userId": "user_001",
    "selectedCharacterId": "char_enfj_001"
  },
  "error": null
}
```

---

# 6. Task API

## 6.1 タスク作成

```http
POST /api/v1/users/{userId}/tasks
```

### Request

```json
{
  "title": "数学の課題を終わらせる"
}
```

### Response

```json
{
  "success": true,
  "data": {
    "taskId": "task_001",
    "userId": "user_001",
    "title": "数学の課題を終わらせる",
    "status": "todo",
    "createdAt": "2026-05-13T12:00:00+09:00",
    "completedAt": null
  },
  "error": null
}
```

---

## 6.2 タスク一覧取得

```http
GET /api/v1/users/{userId}/tasks
```

### Query

```txt
status=todo | done | all
date=2026-05-13
```

### Example

```http
GET /api/v1/users/user_001/tasks?status=all&date=2026-05-13
```

### Response

```json
{
  "success": true,
  "data": [
    {
      "taskId": "task_001",
      "title": "数学の課題を終わらせる",
      "status": "todo",
      "createdAt": "2026-05-13T12:00:00+09:00",
      "completedAt": null
    },
    {
      "taskId": "task_002",
      "title": "英単語を30個覚える",
      "status": "done",
      "createdAt": "2026-05-13T09:00:00+09:00",
      "completedAt": "2026-05-13T10:00:00+09:00"
    }
  ],
  "error": null
}
```

---

## 6.3 タスク完了

```http
PATCH /api/v1/users/{userId}/tasks/{taskId}/complete
```

### Response

```json
{
  "success": true,
  "data": {
    "taskId": "task_001",
    "status": "done",
    "completedAt": "2026-05-13T15:00:00+09:00"
  },
  "error": null
}
```

---

## 6.4 タスク未完了に戻す

```http
PATCH /api/v1/users/{userId}/tasks/{taskId}/reopen
```

### Response

```json
{
  "success": true,
  "data": {
    "taskId": "task_001",
    "status": "todo",
    "completedAt": null
  },
  "error": null
}
```

---

## 6.5 タスク削除

```http
DELETE /api/v1/users/{userId}/tasks/{taskId}
```

### Response

```json
{
  "success": true,
  "data": {
    "taskId": "task_001",
    "deleted": true
  },
  "error": null
}
```

---

# 7. Entertainment Time API

## 7.1 娯楽時間登録・更新

MVPではスクリーンタイム自動取得ではなく、手動入力。

```http
PUT /api/v1/users/{userId}/entertainment-times/{date}
```

### Request

```json
{
  "minutes": 180
}
```

### Response

```json
{
  "success": true,
  "data": {
    "recordId": "ent_001",
    "userId": "user_001",
    "date": "2026-05-13",
    "minutes": 180,
    "targetMinutes": 120,
    "diffMinutes": 60,
    "createdAt": "2026-05-13T20:00:00+09:00",
    "updatedAt": "2026-05-13T20:00:00+09:00"
  },
  "error": null
}
```

---

## 7.2 娯楽時間取得

```http
GET /api/v1/users/{userId}/entertainment-times/{date}
```

### Response

```json
{
  "success": true,
  "data": {
    "recordId": "ent_001",
    "date": "2026-05-13",
    "minutes": 180,
    "targetMinutes": 120,
    "diffMinutes": 60
  },
  "error": null
}
```

---

## 7.3 娯楽時間履歴取得

```http
GET /api/v1/users/{userId}/entertainment-times
```

### Query

```txt
from=2026-05-07
to=2026-05-13
```

### Response

```json
{
  "success": true,
  "data": [
    {
      "date": "2026-05-07",
      "minutes": 100,
      "targetMinutes": 120,
      "diffMinutes": -20
    },
    {
      "date": "2026-05-13",
      "minutes": 180,
      "targetMinutes": 120,
      "diffMinutes": 60
    }
  ],
  "error": null
}
```

---

# 8. Stats API

## 8.1 今日の簡易ステータス取得

チャット画面の上部やサイドに表示する。

```http
GET /api/v1/users/{userId}/stats/today
```

### Response

```json
{
  "success": true,
  "data": {
    "date": "2026-05-13",
    "tasks": {
      "completedCount": 3,
      "todoCount": 2,
      "totalCount": 5,
      "completionRate": 60
    },
    "entertainment": {
      "minutes": 180,
      "targetMinutes": 120,
      "diffMinutes": 60
    },
    "summaryText": "今日は5件中3件のタスクを完了しています。娯楽時間は目標を60分超過しています。"
  },
  "error": null
}
```

---

## 8.2 指定日の統計取得

```http
GET /api/v1/users/{userId}/stats/daily/{date}
```

### Response

```json
{
  "success": true,
  "data": {
    "date": "2026-05-13",
    "completedTaskCount": 3,
    "todoTaskCount": 2,
    "totalTaskCount": 5,
    "taskCompletionRate": 60,
    "entertainmentMinutes": 180,
    "targetEntertainmentMinutes": 120,
    "entertainmentDiffMinutes": 60,
    "summaryText": "タスクは半分以上完了。娯楽時間はやや多めです。"
  },
  "error": null
}
```

---

## 8.3 過去7日分の統計取得

```http
GET /api/v1/users/{userId}/stats/weekly
```

### Query

```txt
endDate=2026-05-13
```

### Response

```json
{
  "success": true,
  "data": {
    "from": "2026-05-07",
    "to": "2026-05-13",
    "days": [
      {
        "date": "2026-05-07",
        "completedTaskCount": 4,
        "todoTaskCount": 1,
        "taskCompletionRate": 80,
        "entertainmentMinutes": 100,
        "targetEntertainmentMinutes": 120,
        "entertainmentDiffMinutes": -20
      }
    ]
  },
  "error": null
}
```

---

# 9. Chat API

## 9.1 チャット送信

ユーザの発言を保存し、AI返答を生成し、必要であればボイス生成まで行う。

```http
POST /api/v1/users/{userId}/chat/messages
```

### Request

```json
{
  "message": "今日はちょっとゲームしすぎたかも",
  "characterId": "char_istj_001",
  "generateVoice": true
}
```

### Backend Internal Flow

```txt
1. ユーザ発言をチャットログに保存
2. 選択中キャラ情報を取得
3. 今日のタスク状況を取得
4. 今日の娯楽時間を取得
5. 直近チャットログを取得
6. AI用プロンプトを組み立て
7. AI返答を生成
8. AI返答をチャットログに保存
9. generateVoice=true の場合、音声生成
10. 返答テキストと音声URLを返す
```

### Response

```json
{
  "success": true,
  "data": {
    "userMessage": {
      "chatLogId": "chat_001",
      "speaker": "user",
      "message": "今日はちょっとゲームしすぎたかも",
      "createdAt": "2026-05-13T21:00:00+09:00"
    },
    "assistantMessage": {
      "chatLogId": "chat_002",
      "speaker": "character",
      "characterId": "char_istj_001",
      "message": "社長、正直に報告できたのは良いことです。ただ、明日は目標時間を意識して、先にタスクを片付けましょう。",
      "voiceUrl": "/api/v1/voice-files/voice_file_001",
      "createdAt": "2026-05-13T21:00:05+09:00"
    },
    "context": {
      "todayTaskCompletedCount": 3,
      "todayTaskTotalCount": 5,
      "todayEntertainmentMinutes": 180,
      "targetEntertainmentMinutes": 120
    }
  },
  "error": null
}
```

---

## 9.2 チャット履歴取得

チャット画面で使用する。

```http
GET /api/v1/users/{userId}/chat/messages
```

### Query

```txt
date=2026-05-13
characterId=char_istj_001
limit=50
cursor=xxx
```

### Response

```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "chatLogId": "chat_001",
        "speaker": "user",
        "message": "今日はちょっとゲームしすぎたかも",
        "characterId": "char_istj_001",
        "voiceUrl": null,
        "createdAt": "2026-05-13T21:00:00+09:00"
      },
      {
        "chatLogId": "chat_002",
        "speaker": "character",
        "message": "社長、正直に報告できたのは良いことです。",
        "characterId": "char_istj_001",
        "voiceUrl": "/api/v1/voice-files/voice_file_001",
        "createdAt": "2026-05-13T21:00:05+09:00"
      }
    ],
    "nextCursor": null
  },
  "error": null
}
```

---

# 10. Voice API

## 10.1 テキストから音声生成

チャットとは別に、再生成したい場合に使う。

```http
POST /api/v1/users/{userId}/voices
```

### Request

```json
{
  "characterId": "char_istj_001",
  "text": "社長、本日のタスク進捗を確認しましょう。"
}
```

### Response

```json
{
  "success": true,
  "data": {
    "voiceFileId": "voice_file_001",
    "characterId": "char_istj_001",
    "text": "社長、本日のタスク進捗を確認しましょう。",
    "voiceUrl": "/api/v1/voice-files/voice_file_001",
    "createdAt": "2026-05-13T12:00:00+09:00"
  },
  "error": null
}
```

---

## 10.2 音声ファイル取得

```http
GET /api/v1/voice-files/{voiceFileId}
```

### Response

音声ファイルを返す。

```txt
Content-Type: audio/mpeg
```

または

```txt
Content-Type: audio/wav
```

---

# 11. Chat Log API

チャット画面用の履歴取得とは別に、ログ画面用のAPIを用意する。

## 11.1 日付別チャットログ取得

```http
GET /api/v1/users/{userId}/chat-logs
```

### Query

```txt
date=2026-05-13
characterId=char_istj_001
```

### Response

```json
{
  "success": true,
  "data": {
    "date": "2026-05-13",
    "characterId": "char_istj_001",
    "logs": [
      {
        "chatLogId": "chat_001",
        "speaker": "user",
        "message": "おはよう",
        "createdAt": "2026-05-13T08:00:00+09:00"
      },
      {
        "chatLogId": "chat_002",
        "speaker": "character",
        "message": "おはようございます、社長。本日も予定を確認しましょう。",
        "voiceUrl": "/api/v1/voice-files/voice_file_001",
        "createdAt": "2026-05-13T08:00:05+09:00"
      }
    ]
  },
  "error": null
}
```

---

## 11.2 チャットログ日付一覧取得

ログ画面で、会話が存在する日付一覧を表示する。

```http
GET /api/v1/users/{userId}/chat-logs/dates
```

### Response

```json
{
  "success": true,
  "data": [
    {
      "date": "2026-05-13",
      "messageCount": 24,
      "characters": [
        {
          "characterId": "char_istj_001",
          "name": "白石 澪"
        }
      ]
    }
  ],
  "error": null
}
```

---

# 12. 初期設定画面で使うAPI

## 12.1 初期設定に必要なデータ取得

```http
GET /api/v1/setup/options
```

### Response

```json
{
  "success": true,
  "data": {
    "characters": [
      {
        "characterId": "char_istj_001",
        "name": "白石 澪",
        "mbti": "ISTJ",
        "description": "真面目で几帳面な管理型秘書",
        "iconUrl": "/assets/characters/istj/icon.png",
        "sampleVoiceUrl": "/assets/voices/istj/sample.wav"
      }
    ],
    "defaultTargetEntertainmentMinutes": 120
  },
  "error": null
}
```

---

# 13. チャット用AIコンテキスト設計

## 13.1 フロントから送る情報

フロントからは最小限でよい。

```json
{
  "message": "今日はゲームしすぎた",
  "characterId": "char_istj_001",
  "generateVoice": true
}
```

## 13.2 バックエンド側で集める情報

バックエンドで以下を取得する。

```json
{
  "user": {
    "name": "太郎",
    "presidentName": "社長"
  },
  "character": {
    "name": "白石 澪",
    "mbti": "ISTJ",
    "description": "真面目で几帳面な管理型秘書",
    "tone": "丁寧で落ち着いた口調"
  },
  "todayStats": {
    "completedTaskCount": 3,
    "todoTaskCount": 2,
    "taskCompletionRate": 60,
    "entertainmentMinutes": 180,
    "targetEntertainmentMinutes": 120,
    "entertainmentDiffMinutes": 60
  },
  "recentChatLogs": [
    {
      "speaker": "user",
      "message": "今日の予定を確認したい"
    },
    {
      "speaker": "character",
      "message": "承知しました、社長。まずはタスク一覧を確認しましょう。"
    }
  ]
}
```

## 13.3 AI返答の制約

AIには以下の制約を渡す。

```txt
- ユーザを社長として扱う
- キャラのMBTI、性格、口調を守る
- タスク状況に必要に応じて触れる
- 娯楽時間に必要に応じて触れる
- 必要以上に責めない
- 返答は短く自然な会話にする
- 医療、法律、金融などの専門助言は避ける
- タスク管理アプリの秘書として振る舞う
```

---

# 14. MVPで優先実装するAPI順

## Phase 1：基盤

```txt
POST /users
GET /users/{userId}
PATCH /users/{userId}
GET /characters
GET /characters/{characterId}
PATCH /users/{userId}/selected-character
```

## Phase 2：タスク

```txt
POST /users/{userId}/tasks
GET /users/{userId}/tasks
PATCH /users/{userId}/tasks/{taskId}/complete
PATCH /users/{userId}/tasks/{taskId}/reopen
DELETE /users/{userId}/tasks/{taskId}
```

## Phase 3：娯楽時間

```txt
PUT /users/{userId}/entertainment-times/{date}
GET /users/{userId}/entertainment-times/{date}
GET /users/{userId}/entertainment-times
```

## Phase 4：統計

```txt
GET /users/{userId}/stats/today
GET /users/{userId}/stats/daily/{date}
GET /users/{userId}/stats/weekly
```

## Phase 5：チャット

```txt
POST /users/{userId}/chat/messages
GET /users/{userId}/chat/messages
GET /users/{userId}/chat-logs
GET /users/{userId}/chat-logs/dates
```

## Phase 6：ボイス

```txt
POST /users/{userId}/voices
GET /voice-files/{voiceFileId}
```

---

# 15. 最初に実装すべき最小APIセット

本当に最小で動かすなら、まずは以下だけでよい。

```txt
POST /api/v1/users
GET /api/v1/characters
PATCH /api/v1/users/{userId}/selected-character

POST /api/v1/users/{userId}/tasks
GET /api/v1/users/{userId}/tasks
PATCH /api/v1/users/{userId}/tasks/{taskId}/complete

PUT /api/v1/users/{userId}/entertainment-times/{date}
GET /api/v1/users/{userId}/stats/today

POST /api/v1/users/{userId}/chat/messages
GET /api/v1/users/{userId}/chat/messages
```

この時点で成立する体験：

- ユーザ登録
- 秘書キャラ選択
- タスク登録
- タスク完了
- 娯楽時間入力
- 今日の統計確認
- キャラとのチャット
- チャットログ保存

---

# 16. 補足：MVPではまだ作らなくてよいAPI

MVP初期では以下は後回しでよい。

```txt
通知API
好感度API
長期記憶API
スクリーンタイム自動取得API
詳細分析API
ランキングAPI
サブスクリプションAPI
複数端末同期API
管理者API
```

---

# 17. API設計上の重要ポイント

## 17.1 チャットAPIに責務を寄せすぎない

チャットAPIは便利だが、以下を全部フロントから送らない。

```txt
タスク状況
娯楽時間
統計
キャラ設定
直近ログ
```

これらはバックエンド側で取得する。

理由：

- フロントの実装が軽くなる
- 改ざんされにくい
- AIプロンプトの形式をバックエンドで管理できる
- 将来、記憶や好感度を追加しやすい

## 17.2 統計は最初は動的計算でよい

MVPでは stats テーブルに全部保存しなくてもよい。

最初は以下から計算する。

```txt
tasks
entertainment_times
users.targetEntertainmentMinutes
```

あとで重くなったら daily_stats を作る。

## 17.3 ボイス生成はチャットAPIに含めてよい

MVPではチャット返信後にボイスが必要なので、以下の形が扱いやすい。

```json
{
  "message": "今日の進捗を確認しましょう。",
  "voiceUrl": "/api/v1/voice-files/voice_file_001"
}
```

**補足（ボイスファイルの管理）**:
ボイスの実体はローカルのファイルシステムやオブジェクトストレージ等に保存し、DBの `chat_logs` テーブルにはその識別子（`voice_file_id`）のみを保持する。フロントエンドからは `/api/v1/voice-files/{voice_file_id}` を通じて音声を取得する。
ただし、音声生成が遅い場合は将来的に非同期化する。

## 17.4 キャラ人格はDBで管理する

キャラごとに以下を持たせる。

```txt
name
mbti
description
tone
voiceId
iconUrl
standingImageUrl
systemPromptFragment
```

特に `systemPromptFragment` を持たせると、キャラ別のAI応答を管理しやすい。

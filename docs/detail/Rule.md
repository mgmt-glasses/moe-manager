/# 開発ルール

## 1. アーキテクチャルール

本プロジェクトは **モジュラーモノリス × ヘキサゴナルアーキテクチャ** を採用している。

### プロジェクト構成

```
moe-manager/
├── apps/
│   └── gateway/          # 各モジュールを束ねる FastAPI アプリケーション
├── packages/
│   ├── user/             # ユーザ管理モジュール
│   ├── character/        # キャラ管理モジュール
│   ├── task/             # タスク管理モジュール
│   ├── screentime/       # スクリーンタイム管理モジュール
│   ├── statistics/       # 統計管理モジュール
│   └── voice-library/    # ボイス・チャットモジュール
└── pyproject.toml        # uv ワークスペース定義
```

### モジュール内のディレクトリ構成

各モジュールはヘキサゴナルアーキテクチャに従い、以下の構成を持つ。

```
packages/<module>/moe_<module>/
├── domain/
│   ├── models.py         # ドメインモデル（Pydantic BaseModel, Enum 等）
│   ├── ports.py          # ポート定義（Protocol による抽象インターフェース）
│   └── use_cases.py      # ユースケース（ビジネスロジック）
└── adapters/
    ├── inbound/
    │   └── api/
    │       └── <module>_router.py   # FastAPI ルーター（外部 → domain への入口）
    └── outbound/
        └── repositories/
            └── postgres_<module>_repository.py  # リポジトリ実装（domain → 外部への出口）
```

### 依存方向のルール

**最重要ルール: 依存の方向は常に「外側 → 内側」。domain は何にも依存しない。**

```
adapter（外側）→ use_cases → ports / models（内側）
```

| レイヤー | 依存してよいもの | 依存してはいけないもの |
|---|---|---|
| `domain/models.py` | 標準ライブラリ、pydantic のみ | ports, use_cases, adapters, 他モジュール |
| `domain/ports.py` | models | use_cases, adapters, 他モジュール |
| `domain/use_cases.py` | models, ports | adapters, 他モジュール |
| `adapters/inbound/` | domain 全体（models, ports, use_cases） | outbound adapters |
| `adapters/outbound/` | domain/models, domain/ports | inbound adapters, use_cases |

### モジュール間の境界ルール

- モジュール間で直接 import しない。モジュールは独立したパッケージとして扱う
- モジュール間の連携が必要な場合は、**gateway 層**で組み合わせるか、**共通のインターフェース（Port）** を介して行う
- あるモジュールの domain が別モジュールの domain を直接参照することは禁止
- 共有が必要な型がある場合は、共通パッケージ（`packages/shared` 等）の作成を検討する

### 新規コードの追加ルール

| 追加するもの | 配置場所 |
|---|---|
| ドメインモデル | `domain/models.py` |
| 抽象インターフェース | `domain/ports.py`（Protocol で定義） |
| ビジネスロジック | `domain/use_cases.py` |
| REST API エンドポイント | `adapters/inbound/api/<module>_router.py` |
| DB アクセス実装 | `adapters/outbound/repositories/` |
| 外部 API 連携 | `adapters/outbound/` 配下に新ディレクトリを作成 |
| 新モジュール | `packages/<module>/` に上記構成で作成し、gateway に router を追加 |

---

## 2. ブランチ戦略

Git Flow ベースのブランチ運用を行う。

### ブランチ構成

| ブランチ | 用途 | マージ先 |
|---|---|---|
| `main` | 本番リリース用。常にデプロイ可能な状態を保つ | - |
| `develop` | 開発統合ブランチ。次のリリースに向けた最新コードを集約する | `main` |
| `feature/*` | 機能開発用。Issue 単位で作成する | `develop` |
| `fix/*` | バグ修正用 | `develop` |
| `hotfix/*` | 本番の緊急修正用 | `main` および `develop` |

### ブランチ命名規則

```
feature/<モジュール名>-<機能名>
fix/<モジュール名>-<修正内容>
hotfix/<修正内容>
```

モジュールをまたぐ変更や基盤変更の場合は、モジュール名の代わりに `gateway` や `infra` を使う。

例：

- `feature/task-crud-api`
- `feature/character-mbti-selection`
- `feature/gateway-router-setup`
- `fix/screentime-date-calculation`
- `hotfix/auth-error`

### 運用ルール

- `main` および `develop` へ直接コミットしない。必ず PR 経由でマージする
- `feature/*` ブランチは `develop` から切り、`develop` へ PR を出す
- マージは **Squash Merge** を基本とする
- マージ後の `feature/*` ブランチはリモートから削除する
- コンフリクトが発生した場合は、`develop` を `feature/*` へマージして解消する（rebase でも可）
- 1 つの PR では **原則 1 モジュール** の変更に閉じる。複数モジュールにまたがる場合は PR の概要にその理由を明記する

---

## 3. コミットメッセージ規則

[Conventional Commits](https://www.conventionalcommits.org/) に準拠する。

### フォーマット

```
<type>(<scope>): <subject>
```

`scope` は省略可。記載する場合は変更対象のモジュール名を入れる。

### type 一覧

| type | 用途 |
|---|---|
| `feat` | 新機能の追加 |
| `fix` | バグ修正 |
| `docs` | ドキュメントの追加・修正 |
| `style` | コードの意味に影響しないフォーマット変更（空白、セミコロン等） |
| `refactor` | リファクタリング（機能追加でもバグ修正でもない変更） |
| `test` | テストの追加・修正 |
| `chore` | ビルド設定、CI、依存パッケージの更新等 |
| `perf` | パフォーマンス改善 |

### scope 一覧

| scope | 対象 |
|---|---|
| `user` | ユーザ管理モジュール |
| `character` | キャラ管理モジュール |
| `task` | タスク管理モジュール |
| `screentime` | スクリーンタイム管理モジュール |
| `statistics` | 統計管理モジュール |
| `voice` | ボイス・チャットモジュール |
| `gateway` | gateway アプリケーション |
| なし | 横断的な変更、ドキュメント等 |

### ルール

- subject は日本語・英語どちらでも可。ただしリポジトリ内で統一する
- subject は簡潔に、何をしたかが分かるように書く
- 本文（body）は必要に応じて空行を挟んで記述する
- 関連 Issue がある場合は末尾に `Refs #<issue-number>` を付ける

### 例

```
feat(task): タスク管理のCRUD APIを追加
```

```
fix(screentime): 日付計算が1日ずれる問題を修正

目標時間との差分計算でタイムゾーンが考慮されていなかった。
Refs #42
```

```
feat(gateway): statisticsモジュールのrouterを統合
```

```
docs: MVP仕様書を追加
```

---

## 4. レビュールール

### PR 作成時のルール

- PR テンプレートに従って記述する
- 1 つの PR では 1 つの目的に絞る（機能追加とリファクタリングを混ぜない）
- 変更行数の目安は **300 行以内**。超える場合は分割を検討する
- Draft PR を活用し、実装途中でも早めにフィードバックをもらう
- 関連 Issue がある場合は PR に紐付ける（`Closes #<issue-number>`）

### レビュー観点

レビュアーは以下の観点で確認する。

- **動作**: 仕様通りに動作するか
- **依存方向**: domain が adapter や外部ライブラリに依存していないか。依存の矢印が内側→外側に向いていないか
- **モジュール境界**: 他モジュールの内部実装を直接 import していないか。モジュール間の結合が生まれていないか
- **Port / Adapter の整合性**: 新しい外部連携を追加する場合、Port（Protocol）が domain に定義され、Adapter が adapters/ に配置されているか
- **可読性**: コードの意図が読み取れるか。不要なコメントや複雑すぎるロジックがないか
- **テスト**: 変更に対する適切なテストがあるか。domain のユースケースは adapter のモックで単体テスト可能か
- **影響範囲**: 既存機能への影響がないか

### レビューの進め方

- PR が作成されたら、チームメンバーは **24 時間以内** にレビューを開始する
- Approve が **1 名以上** でマージ可能とする
- 指摘事項は以下のプレフィックスで分類する

| プレフィックス | 意味 |
|---|---|
| `MUST` | 必ず修正が必要。修正しないとマージ不可 |
| `IMO` | 自分ならこうする、という提案。対応は任意 |
| `NIT` | 細かい指摘（typo、命名等）。対応は任意 |
| `Q` | 質問。理解のための確認 |

### レビュー後の対応

- `MUST` の指摘はすべて対応してから再レビューを依頼する
- `IMO` / `NIT` は対応する・しないをコメントで明示する
- 修正後は該当コメントに返信し、Resolve する

---

## 5. Issue ドキュメント

各 Issue には、作業内容を整理したドキュメントを 1 つ作成し、`docs/Issue/` に保存する。
実装前に背景・スコープ・完了条件を明確にし、レビュー時の判断基準をそろえることを目的とする。

### 保存場所とファイル名

- 保存先は `docs/Issue/`。Issue ごとに 1 ファイル作成する。
- ファイル名は `issue-<Issue番号>-<英小文字スラッグ>.md` とする。
- スラッグはハイフン区切りで、対象と内容が分かる短い英語にする。

例：

```
docs/Issue/issue-12-task-crud-api.md
docs/Issue/issue-23-screentime-input-screen.md
```

### 記入項目

各 Issue ドキュメントには、以下の項目を上から順に記入する。

| 項目 | 内容 |
|---|---|
| タイトル | `# Issue #<番号> <概要>` |
| 背景・目的 | なぜこの Issue が必要か。解決したい課題 |
| 対象モジュール | 変更対象の scope（`user` / `character` / `task` / `screentime` / `statistics` / `voice` / `gateway`）。複数あれば全て記載する |
| スコープ | この Issue で「やること」と「やらないこと」を分けて書く |
| 完了条件 | 受け入れ条件をチェックリストで書く。すべて満たせばクローズ可能 |
| タスク分解 | 実装手順を箇条書きで分解する |
| 関連リンク | 関連 Issue、PR、参照する方針・仕様ドキュメント |

### テンプレート

```markdown
# Issue #<番号> <概要>

## 背景・目的

<なぜ必要か。解決したい課題>

## 対象モジュール

- <scope 名>

## スコープ

### やること
- <...>

### やらないこと
- <...>

## 完了条件

- [ ] <受け入れ条件>
- [ ] <受け入れ条件>

## タスク分解

1. <実装手順>
2. <実装手順>

## 関連リンク

- 関連 Issue: #<番号>
- 関連 PR: #<番号>
- 参照ドキュメント: <パス>
```

### 運用ルール

- Issue を作成したら、同時に Issue ドキュメントも作成する。
- ファイル名の `<Issue番号>` は GitHub の Issue 番号に合わせる。
- ブランチ名（`feature/<モジュール名>-<機能名>`）と Issue ドキュメントの対象モジュールを一致させる。
- 作業中にスコープや完了条件が変わった場合は、Issue ドキュメントも更新する。
- 作業完了時に完了条件のチェックを埋め、関連 PR 番号を追記する。
- Issue ドキュメントの追加・更新は、対応する作業の PR に含める。

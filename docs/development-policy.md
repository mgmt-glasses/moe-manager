# 開発方針

このドキュメントは、Moe Manager のチーム開発で必ず守る基本方針を定義します。
詳細な設計判断は、[アーキテクチャ方針](architecture-guidelines.md) と [実装ガイド](implementation-guide.md) を参照してください。

## 基本方針

- 1 つの PR は 1 つの目的に絞る。
- 原則として、1 つの PR は 1 モジュール内の変更に閉じる。
- 複数モジュールにまたがる場合は、PR の概要に理由と影響範囲を書く。
- domain 層を中心に実装し、外部依存は adapter に閉じ込める。
- 仕様変更、設計判断、運用ルールの変更はドキュメントにも反映する。

## ブランチ運用

Git Flow ベースで運用します。

| ブランチ | 用途 | マージ先 |
| --- | --- | --- |
| `main` | 本番リリース用。常にデプロイ可能な状態を保つ | - |
| `develop` | 開発統合ブランチ。次のリリースに向けた最新コードを集約する | `main` |
| `feature/*` | 機能開発用。Issue 単位で作成する | `develop` |
| `fix/*` | バグ修正用 | `develop` |
| `hotfix/*` | 本番の緊急修正用 | `main`, `develop` |

ブランチ名は以下の形式にします。

```txt
feature/<module>-<feature>
fix/<module>-<description>
hotfix/<description>
```

例:

- `feature/task-crud-api`
- `feature/character-mbti-selection`
- `feature/api-router-setup`
- `fix/screentime-date-calculation`
- `hotfix/auth-error`

## コミットメッセージ

Conventional Commits に準拠します。

```txt
<type>(<scope>): <subject>
```

主な `type` は以下です。

| type | 用途 |
| --- | --- |
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメント |
| `style` | 意味に影響しない整形 |
| `refactor` | 振る舞いを変えない整理 |
| `test` | テスト |
| `chore` | ビルド設定、依存更新など |
| `perf` | パフォーマンス改善 |

`scope` は変更対象のモジュール名を使います。

| scope | 対象 |
| --- | --- |
| `user` | ユーザ管理 |
| `character` | キャラ管理 |
| `task` | タスク管理 |
| `screentime` | スクリーンタイム管理 |
| `statistics` | 統計管理 |
| `chat` | AI チャット・チャットログ |
| `voice` | ボイス生成・音声ファイル |
| `api` | 統合 API・依存性注入 |

`gateway` は旧 Python 構成の scope として扱い、新規 Go 実装では使用しません。

例:

```txt
feat(task): タスク管理のCRUD APIを追加
fix(screentime): 日付計算が1日ずれる問題を修正
docs: MVP仕様書を整理
```

## PR 作成ルール

- `main` と `develop` へ直接コミットしない。
- `feature/*` は `develop` から切り、`develop` に PR を出す。
- マージは Squash Merge を基本とする。
- マージ後の feature ブランチは削除する。
- 変更行数は 300 行以内を目安にする。
- 大きくなる場合は、domain、adapter、API 統合などに分割する。
- 実装途中でも Draft PR を活用して早めにレビューを受ける。

PR には以下を書く。

- 何を変更したか
- なぜ変更したか
- 影響範囲
- 動作確認内容
- 関連 Issue
- レビュアーに特に見てほしい点

## レビュー方針

レビューでは以下を確認します。

- 仕様通りに動作するか
- domain が adapter や外部サービスに依存していないか
- 他モジュールの内部実装を直接 import していないか
- 新しい外部連携に interface が定義されているか
- adapter が適切な `internal/<domain>/adapter/` 配下に置かれているか
- 変更に対するテスト、または動作確認があるか
- 既存機能への影響範囲が説明されているか

レビュー運用は以下を基本とします。

- PR 作成後、チームメンバーは 24 時間以内にレビューを開始する。
- Approve が 1 名以上でマージ可能とする。
- `MUST` 指摘が残っている PR はマージしない。

レビューコメントの分類は以下を使います。

| プレフィックス | 意味 |
| --- | --- |
| `MUST` | マージ前に必ず対応する |
| `IMO` | 提案。対応は任意 |
| `NIT` | 細かい指摘。対応は任意 |
| `Q` | 質問 |

`MUST` はすべて対応してから再レビューを依頼します。
`IMO` と `NIT` は、対応するかどうかをコメントで明示します。

## 品質ゲート

現時点では、リポジトリ共通の lint、型チェック、テストコマンドは未定義です。
PR では最低限、変更箇所に応じた動作確認内容を `.github/pull_request_template.md` の「動作確認」に記載します。

今後 CI や共通コマンドを追加した場合は、以下をこのドキュメントに追記します。

- セットアップ手順
- API サーバー起動コマンド
- テストコマンド
- lint / format / 型チェックコマンド
- マージ前に必須とする CI

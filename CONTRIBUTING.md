# コントリビュートガイド

## ブランチ運用

`develop` から切り、`develop` に PR を出す。

```
feature/<module>-<feature>   # 例: feature/task-crud-api
fix/<module>-<description>   # 例: fix/chat-prompt-format
```

## ローカル確認

PR 作成前に必ず通す：

```bash
make verify   # go vet + go test + go build
```

## コミット規約

[Conventional Commits](https://www.conventionalcommits.org/) に準拠。

```
feat(chat): AI 返答生成を追加
fix(task): 完了日時が UTC にならない問題を修正
docs: README にセットアップ手順を追加
```

有効なスコープ: `user` `character` `task` `screentime` `statistics` `chat` `voice` `api`

## PR ルール

- タイトルに変更概要を書く
- 本文に「変更内容・理由・影響範囲・動作確認」を書く（テンプレートに従う）
- 1 PR は 300 行以内を目安にする
- `MUST` 指摘はすべて対応してから再レビューを依頼する

詳細は [開発方針](docs/development-policy.md) を参照。

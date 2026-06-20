# Moe Manager 開発ドキュメント

このフォルダは、チーム開発開始前に確認する開発方針の入口です。
既存ドキュメントに散らばっている内容を、読む順番と責務が分かるように再整理しています。

## まず読むもの

1. [開発方針](development-policy.md)
   - チーム開発で必ず守る基本ルールをまとめています。
   - ブランチ、PR、レビュー、実装単位の考え方を確認します。

2. [アーキテクチャ方針](architecture-guidelines.md)
   - モジュラーモノリスとヘキサゴナルアーキテクチャの採用理由、依存方向、モジュール境界をまとめています。
   - 新規機能をどこに追加するか判断するときに参照します。

3. [実装ガイド](implementation-guide.md)
   - モジュールごとの実装手順、配置場所、API/DB 方針をまとめています。
   - 実装前に作業対象のモジュールと変更範囲を確認します。

## 必要に応じて読むもの

- [MVP 仕様サマリー](mvp-summary.md)
  - MVP で実現する体験、機能、画面構成の要約です。
- [Issue ドキュメント](Issue/README.md)
  - Issue ごとの作業範囲、受け入れ条件、実装観点を整理しています。
- [API 方針](api-guidelines.md)
  - API 設計の共通ルールとドメイン一覧です。
- [DB 方針](database-guidelines.md)
  - PostgreSQL 採用、テーブル分類、マイグレーション方針です。
- [テスト方針](testing-guidelines.md)
  - fake の配置、service/handler テストの書き方、ドメインごとのテスト実装状況です。
- [ドキュメント運用](documentation-guidelines.md)
  - ドキュメントを増やす・更新するときのルールです。

## 詳細ドキュメントとの対応

各方針ファイルは、以下の詳細ドキュメントをもとに整理しています。
詳細仕様（リクエスト/レスポンス例、カラム定義など）は元ドキュメントを参照してください。
ただし、Python / FastAPI の実装構成は旧設計であり、Go の実装判断には方針ファイルを優先します。

| 方針ファイル | 主な参照元 |
| --- | --- |
| `development-policy.md` | `detail/Rule.md`, `.github/pull_request_template.md` |
| `architecture-guidelines.md` | `detail/ARCHITECTURE.md`, `detail/Rule.md` |
| `implementation-guide.md` | `detail/ARCHITECTURE.md`, `detail/api_design_mvp.md`, `detail/db-design.md` |
| `mvp-summary.md` | `detail/仕様書.md` |
| `api-guidelines.md` | `detail/api_design_mvp.md` |
| `database-guidelines.md` | `detail/db-design.md` |
| `documentation-guidelines.md` | 既存ドキュメント全体の整理方針 |

## 情報の優先順位

1. 開発運用、依存方向、実装言語、配置場所はこのフォルダの方針ファイルを優先する。
2. API の詳細なリクエスト/レスポンス例は `detail/api_design_mvp.md` を参照する。
3. DB の詳細なカラム定義は `detail/db-design.md` を参照する。ただし、テーブル名が矛盾する場合は `database-guidelines.md` の命名整理を優先する。
4. MVP の体験・画面仕様の詳細は `detail/仕様書.md` を参照する。

方針を変更する場合は該当する方針ファイルを同じ PR で更新してください。
旧設計として明記された詳細ドキュメントは、履歴保存のため更新対象外としてよいものとします。

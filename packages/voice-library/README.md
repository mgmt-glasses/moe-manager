# Voice Library (Humanized AI Core)

このフォルダは、`mkh-voice` プロジェクトから「音声生成機能」と「セリフ（人格・心情）機能」を抽出したものです。
他プロジェクトへ流用する際は、この `voice-library` フォルダをそのままコピーして使用してください。

## 構成
- `mkh_voice/`: コアロジック (Domain, Adapters)
- `data/`: 人格設定 (`personas.json`)、キャラクターボイス設定 (`characters.json`)、参照音声 (`references/`)
- `models/`: 学習済みモデル（HuggingFaceから自動ダウンロードされる場合は空です）

## 導入方法（他レポジトリ）

1. `voice-library` フォルダを移行先レポジトリのルートに配置します。
2. 移行先プロジェクトの `pyproject.toml` に、本ライブラリの依存関係を追加してください。
3. `uv sync` を実行して環境を構築します。

## 主要なエントリーポイント
- **音声生成**: `mkh_voice.domain.use_cases.VoiceGenerationUseCase`
- **対話・心情推論**: `mkh_voice.domain.chat_use_cases.ChatUseCase`

## 注意事項
- SQLite データベース（`chat.db` 等）は含まれていません。初回実行時に自動的に作成されます。
- インポートパスは `mkh_voice` で始まります。必要に応じて置換してください。

# CLAUDE.md

このファイルはこのリポジトリでコードを扱う際の Claude Code へのガイダンスを提供します。

## コマンド

### ビルドと開発
- `go install` - プロバイダーをビルドして `$GOBIN` にインストール
- `go generate` - Terraform サンプルをフォーマットしてドキュメントを生成
- `go build ./...` - ビルド確認のみ

### テスト
- `make testacc` - アクセプタンステストを実行（実際の Langfuse API キーが必要）
- `TF_ACC=1 go test ./... -v -timeout 120m` - 全アクセプタンステストを実行

## ローカル開発環境セットアップ

`~/.terraformrc` に以下の設定が必要（追加済み）:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/kenzo0107/langfuse" = "/Users/kenzo.tanaka/bin"
  }
  direct {}
}
```

`go install` 後、`terraform validate` で動作確認可能。

## アーキテクチャ

Terraform Plugin Framework を使用した Langfuse 用 Terraform プロバイダー。

### コア構造
- `main.go` - エントリーポイント
- `langfuse/` - Langfuse API HTTP クライアント
  - `client.go` - Basic Auth による HTTP クライアント基盤
  - `project.go` - プロジェクト CRUD 操作
- `internal/provider/` - メインプロバイダー実装
  - `provider.go` - プロバイダー設定、リソースとデータソースの登録
  - `project_resource.go` - `langfuse_project` リソース
  - `project_data_source.go` - `langfuse_project` データソース
- `examples/` - Terraform 設定例

### 認証
- `LANGFUSE_PUBLIC_KEY` 環境変数または `public_key` プロバイダー設定
- `LANGFUSE_SECRET_KEY` 環境変数または `secret_key` プロバイダー設定
- `LANGFUSE_HOST` 環境変数または `host` プロバイダー設定（デフォルト: `https://cloud.langfuse.com`）

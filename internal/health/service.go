package health

import "context"

// Status は通信方式に依存しないヘルスチェック結果を表します。
type Status struct {
	// Status はサービスの状態を表す機械判定可能な文字列です。
	Status string
}

// Service は業務ドメインではなく、サービスの稼働状態を判定する運用機能です。
// Livenessは外部サービスやデータベースへ依存させません。
type Service struct{}

// NewService はヘルスチェック用のServiceを構築します。
func NewService() *Service {
	return &Service{}
}

// Liveness はプロセスがリクエストを処理できる状態かを返します。
// 外部依存の状態確認は行わず、readinessとは明確に分離します。
func (s *Service) Liveness(context.Context) Status {
	return Status{Status: "ok"}
}

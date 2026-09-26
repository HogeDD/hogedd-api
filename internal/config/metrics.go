package config

import "os"

// LoadMetricsIngestToken はBFFからの匿名計測を認証する秘密値を返します。
// 未設定時は計測endpointだけがfail closedになります。
func LoadMetricsIngestToken() string { return os.Getenv("METRICS_INGEST_TOKEN") }

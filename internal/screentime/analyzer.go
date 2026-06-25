package screentime

import (
	"context"
)

type CategoryUsage struct {
	Category string `json:"category"`
	Minutes  int    `json:"minutes"`
}

type AnalysisResult struct {
	Items        []CategoryUsage `json:"items"`
	TotalMinutes int             `json:"total_minutes"`
}

type ImageAnalyzer interface {
	AnalyzeImage(ctx context.Context, base64Data, mimeType string) (*AnalysisResult, error)
	Close() error
}

package testutil

import (
	"context"

	"github.com/mgmt-glasses/moe-manager/internal/screentime"
)

type FakeAnalyzer struct {
	Result *screentime.AnalysisResult
	Err    error
	Closed bool
}

func (f *FakeAnalyzer) AnalyzeImage(ctx context.Context, base64Data, mimeType string) (*screentime.AnalysisResult, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Result, nil
}

func (f *FakeAnalyzer) Close() error {
	f.Closed = true
	return nil
}

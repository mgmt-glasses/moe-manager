package adapter_test

import (
	"context"
	"os"
	"testing"

	"github.com/mgmt-glasses/moe-manager/internal/screentime/adapter"
)

func TestGeminiAnalyzer(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY is not set. Skipping integration test.")
	}

	ctx := context.Background()
	analyzer, err := adapter.NewGeminiImageAnalyzer(ctx, apiKey)
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}
	defer analyzer.Close()

	// 1x1 pixel JPEG base64 (dummy)
	dummyJPEG := "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////wgALCAABAAEBAREA/8QAFBABAAAAAAAAAAAAAAAAAAAAAP/aAAgBAQABPxA="

	result, err := analyzer.AnalyzeImage(ctx, dummyJPEG, "image/jpeg")
	
	// Gemini might return an error or empty result because it's a 1x1 white pixel,
	// but we just want to ensure it doesn't PANIC and handles it gracefully.
	t.Logf("Result: %+v, Error: %v", result, err)
}

package adapter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/mgmt-glasses/moe-manager/internal/screentime"
	"google.golang.org/api/option"
)

type GeminiImageAnalyzer struct {
	client *genai.Client
}

func NewGeminiImageAnalyzer(ctx context.Context, apiKey string) (*GeminiImageAnalyzer, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &GeminiImageAnalyzer{client: client}, nil
}

func (a *GeminiImageAnalyzer) Close() error {
	if a.client != nil {
		a.client.Close()
	}
	return nil
}

func (a *GeminiImageAnalyzer) AnalyzeImage(ctx context.Context, base64Data, mimeType string) (*screentime.AnalysisResult, error) {
	// Remove data URI scheme prefix if present
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 image data: %w", err)
	}

	model := a.client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	prompt := genai.Text("スマートフォンのスクリーンタイム画面の画像です。各アプリまたはジャンルの利用時間（分）を抽出してください。JSON配列の形式で、キーは category (アプリ名またはジャンル名), minutes (分数: 整数) にしてください。")
	
	resp, err := model.GenerateContent(ctx, genai.ImageData(strings.TrimPrefix(mimeType, "image/"), decodedBytes), prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content from gemini: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	part := resp.Candidates[0].Content.Parts[0]
	textPart, ok := part.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response type from gemini")
	}

	var items []screentime.CategoryUsage
	if err := json.Unmarshal([]byte(string(textPart)), &items); err != nil {
		return nil, fmt.Errorf("failed to parse json response: %w, response: %s", err, string(textPart))
	}

	total := 0
	for _, item := range items {
		total += item.Minutes
	}

	return &screentime.AnalysisResult{
		Items:        items,
		TotalMinutes: total,
	}, nil
}

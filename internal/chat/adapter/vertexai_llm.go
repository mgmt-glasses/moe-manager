package chatadapter

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/genai"

	"github.com/mgmt-glasses/moe-manager/internal/chat"
)

const (
	cacheTTL    = 55 * time.Minute
	cacheMinTTL = 5 * time.Minute
)

type cacheEntry struct {
	cacheID   string
	expiresAt time.Time
}

type VertexAILLMClient struct {
	client *genai.Client
	model  string
	mu     sync.Mutex
	caches map[string]cacheEntry
}

func NewVertexAILLMClient(ctx context.Context, project, location, model string) (*VertexAILLMClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  project,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("create vertex ai client: %w", err)
	}
	return &VertexAILLMClient{
		client: client,
		model:  model,
		caches: make(map[string]cacheEntry),
	}, nil
}

func (c *VertexAILLMClient) GenerateReply(ctx context.Context, prompt chat.Prompt) (string, error) {
	cacheID, err := c.getOrCreateCache(ctx, prompt.System)
	if err != nil {
		log.Printf("vertexai: cache unavailable (fallback to direct): %v", err)
		return c.generateDirect(ctx, prompt)
	}
	return c.generateWithCache(ctx, cacheID, prompt.User)
}

func (c *VertexAILLMClient) getOrCreateCache(ctx context.Context, systemInstruction string) (string, error) {
	key := systemHash(systemInstruction)

	c.mu.Lock()
	entry, ok := c.caches[key]
	c.mu.Unlock()

	if ok && time.Now().Before(entry.expiresAt.Add(-cacheMinTTL)) {
		return entry.cacheID, nil
	}

	cacheID, err := c.createCache(ctx, systemInstruction)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.caches[key] = cacheEntry{
		cacheID:   cacheID,
		expiresAt: time.Now().Add(cacheTTL),
	}
	c.mu.Unlock()

	log.Printf("vertexai: created context cache %s", cacheID)
	return cacheID, nil
}

func (c *VertexAILLMClient) createCache(ctx context.Context, systemInstruction string) (string, error) {
	cached, err := c.client.Caches.Create(ctx, c.model, &genai.CreateCachedContentConfig{
		SystemInstruction: genai.NewContentFromText(systemInstruction, ""),
		TTL:               cacheTTL,
	})
	if err != nil {
		return "", fmt.Errorf("create cache: %w", err)
	}
	return cached.Name, nil
}

func (c *VertexAILLMClient) generateWithCache(ctx context.Context, cacheID, userMessage string) (string, error) {
	resp, err := c.client.Models.GenerateContent(ctx, c.model,
		[]*genai.Content{genai.NewContentFromText(userMessage, "user")},
		&genai.GenerateContentConfig{
			CachedContent: cacheID,
		},
	)
	if err != nil {
		return "", fmt.Errorf("generate with cache: %w", err)
	}
	return resp.Text(), nil
}

func (c *VertexAILLMClient) generateDirect(ctx context.Context, prompt chat.Prompt) (string, error) {
	resp, err := c.client.Models.GenerateContent(ctx, c.model,
		[]*genai.Content{genai.NewContentFromText(prompt.User, "user")},
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(prompt.System, ""),
		},
	)
	if err != nil {
		return "", fmt.Errorf("generate direct: %w", err)
	}
	return resp.Text(), nil
}

func systemHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}

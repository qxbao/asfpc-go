package generative

import (
	"testing"
)

func TestGenerateText(t *testing.T) {
	service := GetGenerativeService("", "gemini-2.5-flash")
	
	err := service.Init()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}
	
	text, err := service.GenerateText("Say a short hello")
	if err != nil {
		t.Fatalf("Failed to generate text: %v", err)
	} else {
		t.Logf("Generated text: %s", text)
	}
}

func TestEmbedding(t *testing.T) {
	service := GetGenerativeService("", "gemini-embedding-001")
	err := service.Init()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}
	embeddings, err := service.GenerateEmbedding("Test data")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}
	t.Logf("Generated embeddings: %v", len(embeddings))
}
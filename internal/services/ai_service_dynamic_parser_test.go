package services

import "testing"

func TestParseDynamicScenarioResponse_WithUnterminatedMarkdownFence(t *testing.T) {
	input := "```json\n{\n  \"question\": \"Because of your previous decision to rapidly scale features based on initial positive feedback from early adopters\",\n  \"options\": [\n    {\"text\":\"Option A\",\"proficiency\":1,\"feedback\":\"Low\"},\n    {\"text\":\"Option B\",\"proficiency\":2,\"feedback\":\"Mid\"},\n    {\"text\":\"Option C\",\"proficiency\":3,\"feedback\":\"High\"},\n    {\"text\":\"Option D\",\"proficiency\":2,\"feedback\":\"Balanced\"}\n  ]\n}"

	parsed, err := parseDynamicScenarioResponse(input)
	if err != nil {
		t.Fatalf("expected parser to handle fenced JSON, got error: %v", err)
	}

	if parsed.Question == "" {
		t.Fatalf("expected non-empty question")
	}

	if len(parsed.Options) != 4 {
		t.Fatalf("expected 4 options, got %d", len(parsed.Options))
	}
}

func TestParseDynamicScenarioResponse_TruncatedJSONFails(t *testing.T) {
	input := "```json\n{\n  \"question\": \"Because of your previous decision to rapidly scale features\""

	_, err := parseDynamicScenarioResponse(input)
	if err == nil {
		t.Fatalf("expected error for truncated JSON")
	}
}

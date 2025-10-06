package prompt

import (
	"testing"
)

func TestReplacePrompt(t *testing.T) {
	ps := &PromptService{}
	tests := []struct {
		name     string
		prompt   string
		kwargs   []string
		expected string
	}{
		{
			name:     "Basic replacement",
			prompt:   "Hello INSERT_1, welcome to INSERT_2!",
			kwargs:   []string{"Alice", "Wonderland"},
			expected: "Hello Alice, welcome to Wonderland!",
		},
		{
			name:     "Null replacement",
			prompt:   "Hello INSERT_1, your score is INSERT_2.",
			kwargs:   []string{"Bob", ""},
			expected: "Hello Bob, your score is (null).",
		},
		{
			name:     "No placeholders",
			prompt:   "Just a static prompt.",
			kwargs:   []string{"Unused"},
			expected: "Just a static prompt.",
		},
		{
			name:     "Excess kwargs",
			prompt:   "INSERT_1 and INSERT_2 are friends.",
			kwargs:   []string{"Charlie", "Delta", "Extra"},
			expected: "Charlie and Delta are friends.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ps.ReplacePrompt(tt.prompt, tt.kwargs...)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
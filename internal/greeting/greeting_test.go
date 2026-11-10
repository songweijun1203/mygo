package greeting

import "testing"

func TestMessage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "name", input: "Gopher", want: "Hello, Gopher!"},
		{name: "trim spaces", input: "  Go  ", want: "Hello, Go!"},
		{name: "blank name", input: "  ", want: "Hello, Go!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Message(tt.input); got != tt.want {
				t.Fatalf("Message(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

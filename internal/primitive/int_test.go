package primitive

import (
	"testing"
)

func TestIntWrapPrompt(t *testing.T) {
	h := &Int[int]{}
	got := h.WrapPrompt("Count to 10")
	want := `Count to 10

Provide your response as a single integer number.
Do not include any units, symbols, or additional text.
Examples: 42, -17, 0`
	if got != want {
		t.Errorf("WrapPrompt() = %q, want %q", got, want)
	}
}

func TestIntParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "simple number",
			input:   "42",
			want:    42,
			wantErr: false,
		},
		{
			name:    "negative number",
			input:   "-17",
			want:    -17,
			wantErr: false,
		},
		{
			name:    "zero",
			input:   "0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "with whitespace",
			input:   "  123  ",
			want:    123,
			wantErr: false,
		},
		{
			name:    "invalid - decimal",
			input:   "3.14",
			wantErr: true,
		},
		{
			name:    "invalid - text",
			input:   "not a number",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			input:   "",
			wantErr: true,
		},
	}

	h := &Int[int]{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntValidate(t *testing.T) {
	t.Run("valid type", func(t *testing.T) {
		h := &Int[int]{}
		if err := h.Validate(42); err != nil {
			t.Errorf("Validate(42) error = %v, want nil", err)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		h := &Int[string]{}
		var s string = "not an int"
		if err := h.Validate(s); err == nil {
			t.Error("Validate() should fail with wrong type")
		}
	})
}

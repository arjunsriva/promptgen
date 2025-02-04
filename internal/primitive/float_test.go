package primitive

import (
	"testing"
)

func TestFloatWrapPrompt(t *testing.T) {
	h := &Float[float64]{}
	got := h.WrapPrompt("What's pi?")
	want := `What's pi?

Provide your response as a single decimal number.
Use a period (.) as the decimal separator.
Do not include any units, symbols, or additional text.
Examples: 3.14, -2.5, 0.0, 42.0`
	if got != want {
		t.Errorf("WrapPrompt() = %q, want %q", got, want)
	}
}

func TestFloatParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{
			name:    "simple decimal",
			input:   "3.14",
			want:    3.14,
			wantErr: false,
		},
		{
			name:    "negative decimal",
			input:   "-2.5",
			want:    -2.5,
			wantErr: false,
		},
		{
			name:    "whole number",
			input:   "42",
			want:    42.0,
			wantErr: false,
		},
		{
			name:    "zero",
			input:   "0.0",
			want:    0.0,
			wantErr: false,
		},
		{
			name:    "with whitespace",
			input:   "  123.456  ",
			want:    123.456,
			wantErr: false,
		},
		{
			name:    "scientific notation",
			input:   "1.23e-4",
			want:    0.000123,
			wantErr: false,
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
		{
			name:    "invalid - multiple dots",
			input:   "1.2.3",
			wantErr: true,
		},
	}

	h := &Float[float64]{}
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

func TestFloatValidate(t *testing.T) {
	t.Run("valid type", func(t *testing.T) {
		h := &Float[float64]{}
		if err := h.Validate(3.14); err != nil {
			t.Errorf("Validate(3.14) error = %v, want nil", err)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		h := &Float[string]{}
		var s string = "not a float"
		if err := h.Validate(s); err == nil {
			t.Error("Validate() should fail with wrong type")
		}
	})
}

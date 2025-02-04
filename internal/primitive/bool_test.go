package primitive

import (
	"testing"
)

func TestBoolWrapPrompt(t *testing.T) {
	h := &Bool[bool]{}
	got := h.WrapPrompt("Is this correct?")
	want := `Is this correct?

Provide your response as a single word: true or false.
Do not include any additional text or explanation.
Examples: true, false`
	if got != want {
		t.Errorf("WrapPrompt() = %q, want %q", got, want)
	}
}

func TestBoolParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{
			name:    "true",
			input:   "true",
			want:    true,
			wantErr: false,
		},
		{
			name:    "false",
			input:   "false",
			want:    false,
			wantErr: false,
		},
		{
			name:    "yes",
			input:   "yes",
			want:    true,
			wantErr: false,
		},
		{
			name:    "no",
			input:   "no",
			want:    false,
			wantErr: false,
		},
		{
			name:    "1",
			input:   "1",
			want:    true,
			wantErr: false,
		},
		{
			name:    "0",
			input:   "0",
			want:    false,
			wantErr: false,
		},
		{
			name:    "TRUE uppercase",
			input:   "TRUE",
			want:    true,
			wantErr: false,
		},
		{
			name:    "with whitespace",
			input:   "  true  ",
			want:    true,
			wantErr: false,
		},
		{
			name:    "invalid - text",
			input:   "not a boolean",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid - number",
			input:   "42",
			wantErr: true,
		},
	}

	h := &Bool[bool]{}
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

func TestBoolValidate(t *testing.T) {
	t.Run("valid type", func(t *testing.T) {
		h := &Bool[bool]{}
		if err := h.Validate(true); err != nil {
			t.Errorf("Validate(true) error = %v, want nil", err)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		h := &Bool[string]{}
		var s string = "not a bool"
		if err := h.Validate(s); err == nil {
			t.Error("Validate() should fail with wrong type")
		}
	})
}

package primitive

import "testing"

func TestStringHandler(t *testing.T) {
	handler, err := NewString[string]()
	if err != nil {
		t.Fatalf("NewString() error = %v", err)
	}

	t.Run("wrap prompt", func(t *testing.T) {
		prompt := handler.WrapPrompt("Tell me a joke")
		if prompt == "Tell me a joke" {
			t.Error("prompt should be wrapped with instructions")
		}
	})

	t.Run("parse valid string", func(t *testing.T) {
		result, err := handler.Parse("Hello world")
		if err != nil {
			t.Errorf("Parse() error = %v", err)
		}
		if result != "Hello world" {
			t.Errorf("Parse() = %q, want %q", result, "Hello world")
		}
	})

	t.Run("parse with whitespace", func(t *testing.T) {
		result, err := handler.Parse("  Hello  world  \n")
		if err != nil {
			t.Errorf("Parse() error = %v", err)
		}
		if result != "Hello  world" {
			t.Errorf("Parse() = %q, want %q", result, "Hello  world")
		}
	})

	t.Run("validate valid string", func(t *testing.T) {
		if err := handler.Validate("Hello"); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("validate empty string", func(t *testing.T) {
		if err := handler.Validate(""); err == nil {
			t.Error("Validate() should error on empty string")
		}
	})
}

func TestStringHandlerWithWrongType(t *testing.T) {
	handler, err := NewString[int]()
	if err != nil {
		t.Fatalf("NewString() error = %v", err)
	}

	t.Run("parse wrong type", func(t *testing.T) {
		_, err := handler.Parse("123")
		if err == nil {
			t.Error("Parse() should error when output type is not string")
		}
	})

	t.Run("validate wrong type", func(t *testing.T) {
		if err := handler.Validate(123); err == nil {
			t.Error("Validate() should error when output type is not string")
		}
	})
}

package primitive

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("string type", func(t *testing.T) {
		handler, err := New[string]()
		if err != nil {
			t.Errorf("New[string]() error = %v", err)
			return
		}
		if _, ok := handler.(*String[string]); !ok {
			t.Errorf("New[string]() returned wrong type = %T", handler)
		}
	})

	t.Run("int type", func(t *testing.T) {
		handler, err := New[int]()
		if err != nil {
			t.Errorf("New[int]() error = %v", err)
			return
		}
		if _, ok := handler.(*Int[int]); !ok {
			t.Errorf("New[int]() returned wrong type = %T", handler)
		}
	})

	t.Run("float type", func(t *testing.T) {
		handler, err := New[float64]()
		if err != nil {
			t.Errorf("New[float64]() error = %v", err)
			return
		}
		if _, ok := handler.(*Float[float64]); !ok {
			t.Errorf("New[float64]() returned wrong type = %T", handler)
		}
	})

	t.Run("bool type", func(t *testing.T) {
		handler, err := New[bool]()
		if err != nil {
			t.Errorf("New[bool]() error = %v", err)
			return
		}
		if _, ok := handler.(*Bool[bool]); !ok {
			t.Errorf("New[bool]() returned wrong type = %T", handler)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		type TestStruct struct {
			Field string
		}
		_, err := New[TestStruct]()
		if err == nil {
			t.Error("New[struct]() should return error for unsupported type")
		}
	})
}

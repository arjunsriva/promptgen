package handler

import "testing"

type TestStruct struct {
	Field string
}

func TestDetermineType(t *testing.T) {
	tests := []struct {
		name     string
		testType func() Type
		want     Type
	}{
		{
			name: "string type",
			testType: func() Type {
				return DetermineType[string]()
			},
			want: TypeString,
		},
		{
			name: "struct type",
			testType: func() Type {
				return DetermineType[TestStruct]()
			},
			want: TypeJSON,
		},
		{
			name: "complex struct type",
			testType: func() Type {
				type ComplexStruct struct {
					Field1 string
					Field2 int
					Field3 []string
				}
				return DetermineType[ComplexStruct]()
			},
			want: TypeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.testType(); got != tt.want {
				t.Errorf("DetermineType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeString(t *testing.T) {
	tests := []struct {
		typ  Type
		want string
	}{
		{TypeUnknown, "unknown"},
		{TypeString, "string"},
		{TypeJSON, "json"},
		{Type(99), "unknown type 99"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("Type.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

type TestStruct struct {
	Required   string   `json:"required_field" jsonschema:"required,title=Required Field,description=This field is required"`
	Optional   string   `json:"optional_field,omitempty" jsonschema:"maxLength=10,title=Optional Field"`
	MinMax     int      `json:"min_max,omitempty" jsonschema:"minimum=1,maximum=100"`
	Enumerated string   `json:"enum_field,omitempty" jsonschema:"enum=one,enum=two,enum=three"`
	Tags       []string `json:"tags,omitempty" jsonschema:"minItems=1,maxItems=5"`
}

// Test case for Product Copy Generation from README
type ProductInput struct {
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	MainFeatures []string `json:"main_features"`
}

type ProductCopy struct {
	Title       string `json:"title" jsonschema:"required,maxLength=60,description=SEO title length"`
	Description string `json:"description" jsonschema:"required,minLength=50,maxLength=160,description=Meta description length"`
}

// Test case for Review Summary from README
type ReviewInput struct {
	Content   string `json:"content"`
	MaxLength int    `json:"max_length,omitempty"`
}

type ReviewSummary struct {
	Summary   string `json:"summary" jsonschema:"required,maxLength=150"`
	Sentiment string `json:"sentiment" jsonschema:"required,enum=positive,enum=negative,enum=neutral"`
}

// Add this at the top with other test structs
type ValidationFeatures struct {
	// String validations
	Title       string `json:"title" jsonschema:"required,minLength=3,maxLength=50,pattern=^[A-Za-z].*"`
	Description string `json:"description,omitempty" jsonschema:"maxLength=1000"`

	// Numeric validations
	Age   int     `json:"age" jsonschema:"minimum=0,maximum=150"`
	Score float64 `json:"score" jsonschema:"minimum=0,maximum=100,multipleOf=0.5"`

	// Array validations
	Tags   []string `json:"tags" jsonschema:"minItems=1,maxItems=10,uniqueItems=true"`
	Scores []int    `json:"scores" jsonschema:"minItems=1,maxItems=5"`

	// Enum validation
	Status string `json:"status" jsonschema:"enum=pending,enum=active,enum=completed"`

	// Object validation
	Metadata map[string]string `json:"metadata,omitempty" jsonschema:"additionalProperties=true"`
}

func TestSchemaGeneration(t *testing.T) {
	validator, err := NewValidator[TestStruct]()
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}

	schema, err := validator.SchemaString()
	if err != nil {
		t.Fatalf("SchemaString failed: %v", err)
	}

	t.Logf("Generated Schema:\n%s", schema)
}

func TestProductCopySchema(t *testing.T) {
	validator, err := NewValidator[ProductCopy]()
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}

	schema, err := validator.SchemaString()
	if err != nil {
		t.Fatalf("SchemaString failed: %v", err)
	}

	// Parse the schema to verify its structure
	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &schemaMap); err != nil {
		t.Fatalf("invalid schema JSON: %v", err)
	}

	// Verify key requirements
	properties := schemaMap["properties"].(map[string]interface{})

	// Check Title constraints
	title := properties["title"].(map[string]interface{})
	if title["maxLength"].(float64) != 60 {
		t.Errorf("title maxLength should be 60")
	}

	// Check Description constraints
	desc := properties["description"].(map[string]interface{})
	if desc["minLength"].(float64) != 50 || desc["maxLength"].(float64) != 160 {
		t.Errorf("description should have minLength=50 and maxLength=160")
	}

	t.Logf("Product Copy Schema:\n%s", schema)
}

func TestReviewSummarySchema(t *testing.T) {
	validator, err := NewValidator[ReviewSummary]()
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}

	schema, err := validator.SchemaString()
	if err != nil {
		t.Fatalf("SchemaString failed: %v", err)
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &schemaMap); err != nil {
		t.Fatalf("invalid schema JSON: %v", err)
	}

	properties := schemaMap["properties"].(map[string]interface{})

	// Check Summary constraints
	summary := properties["summary"].(map[string]interface{})
	if summary["maxLength"].(float64) != 150 {
		t.Errorf("summary maxLength should be 150")
	}

	// Check Sentiment constraints
	sentiment := properties["sentiment"].(map[string]interface{})
	enum := sentiment["enum"].([]interface{})
	expectedValues := []string{"positive", "negative", "neutral"}
	for i, v := range expectedValues {
		if enum[i].(string) != v {
			t.Errorf("sentiment enum should contain %s", v)
		}
	}

	t.Logf("Review Summary Schema:\n%s", schema)
}

func TestProductCopyValidation(t *testing.T) {
	validator, err := NewValidator[ProductCopy]()
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid product copy",
			json: `{
				"title": "Ergonomic Office Chair with Lumbar Support",
				"description": "Experience ultimate comfort with this premium ergonomic office chair featuring adjustable lumbar support and breathable mesh back design."
			}`,
			wantErr: false,
		},
		{
			name: "title too long",
			json: `{
				"title": "Super Ultra Premium Deluxe Ergonomic Office Chair with Advanced Lumbar Support System and Adjustable Features",
				"description": "A good description that meets the length requirements."
			}`,
			wantErr: true,
		},
		{
			name: "description too short",
			json: `{
				"title": "Good Title",
				"description": "Too short"
			}`,
			wantErr: true,
		},
		{
			name: "missing required field",
			json: `{
				"title": "Good Title"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReviewSummaryValidation(t *testing.T) {
	validator, err := NewValidator[ReviewSummary]()
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid review summary",
			json: `{
				"summary": "Customer loved the product's durability and ease of use. Highly satisfied with purchase.",
				"sentiment": "positive"
			}`,
			wantErr: false,
		},
		{
			name: "summary too long",
			json: `{
				"summary": "This summary is way too long and exceeds the maximum length of 150 characters. It goes on and on with unnecessary details that should have been omitted to make it more concise and to the point but fails to do so.",
				"sentiment": "positive"
			}`,
			wantErr: true,
		},
		{
			name: "invalid sentiment",
			json: `{
				"summary": "Good summary",
				"sentiment": "mixed"
			}`,
			wantErr: true,
		},
		{
			name: "missing required fields",
			json: `{
				"summary": "Good summary"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				t.Logf("Error: %v", err)
			}
		})
	}
}

func TestValidationFeatures(t *testing.T) {
	validator, err := NewValidator[ValidationFeatures]()
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid full object",
			json: `{
				"title": "Valid Title",
				"description": "A valid description",
				"age": 25,
				"score": 95.5,
				"tags": ["tag1", "tag2"],
				"scores": [1, 2, 3],
				"status": "active",
				"metadata": {"key": "value"}
			}`,
			wantErr: false,
		},
		{
			name:    "title too short",
			json:    `{"title": "A", "tags": ["tag1"], "status": "active"}`,
			wantErr: true,
			errMsg:  "String length must be greater than or equal to 3",
		},
		{
			name:    "invalid age",
			json:    `{"title": "Valid Title", "age": -1, "tags": ["tag1"], "status": "active"}`,
			wantErr: true,
			errMsg:  "Must be greater than or equal to 0",
		},
		{
			name:    "duplicate tags",
			json:    `{"title": "Valid Title", "tags": ["tag1", "tag1"], "status": "active"}`,
			wantErr: true,
			errMsg:  "array items[0,1] must be unique",
		},
		{
			name:    "invalid status",
			json:    `{"title": "Valid Title", "tags": ["tag1"], "status": "invalid"}`,
			wantErr: true,
			errMsg:  "must be one of the following",
		},
		{
			name:    "invalid score multiple",
			json:    `{"title": "Valid Title", "tags": ["tag1"], "status": "active", "score": 95.7}`,
			wantErr: true,
			errMsg:  "Must be a multiple of 0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate([]byte(tt.json))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %v", tt.errMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

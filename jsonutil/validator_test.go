package jsonutil

import (
	"testing"
)

type TestStruct struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Age      int    `json:"age" validate:"required,min=18,max=100"`
	Username string `json:"username" validate:"required,min=3,max=20"`
}

type OptionalFieldsStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"omitempty,email"`
}

func TestDefaultValidator_Validate_Success(t *testing.T) {
	validator := DefaultValidator()

	valid := &TestStruct{
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
		Username: "testuser",
	}

	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("expected no error for valid struct, got: %v", err)
	}
}

func TestDefaultValidator_Validate_RequiredFields(t *testing.T) {
	validator := DefaultValidator()

	invalid := &TestStruct{
		Email:    "",
		Password: "",
		Age:      0,
		Username: "",
	}

	err := validator.Validate(invalid)
	if err == nil {
		t.Fatal("expected validation error for empty required fields")
	}

	validationErrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	expectedFields := []string{"email", "password", "age", "username"}
	for _, field := range expectedFields {
		if _, exists := validationErrs[field]; !exists {
			t.Errorf("expected validation error for field '%s'", field)
		}
	}
}

func TestDefaultValidator_Validate_EmailFormat(t *testing.T) {
	validator := DefaultValidator()

	tests := []struct {
		name        string
		email       string
		shouldError bool
	}{
		{"valid email", "user@example.com", false},
		{"invalid email - no @", "userexample.com", true},
		{"invalid email - no domain", "user@", true},
		{"invalid email - no user", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{
				Email:    tt.email,
				Password: "password123",
				Age:      25,
				Username: "testuser",
			}

			err := validator.Validate(s)
			if tt.shouldError && err == nil {
				t.Error("expected validation error for invalid email")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("expected no error for valid email, got: %v", err)
			}

			if tt.shouldError && err != nil {
				validationErrs, ok := err.(ValidationErrors)
				if !ok {
					t.Fatalf("expected ValidationErrors, got %T", err)
				}
				if _, exists := validationErrs["email"]; !exists {
					t.Error("expected validation error for 'email' field")
				}
			}
		})
	}
}

func TestDefaultValidator_Validate_MinMaxConstraints(t *testing.T) {
	validator := DefaultValidator()

	tests := []struct {
		name        string
		password    string
		age         int
		username    string
		shouldError bool
		errorField  string
	}{
		{"password too short", "short", 25, "testuser", true, "password"},
		{"age too young", "password123", 17, "testuser", true, "age"},
		{"age too old", "password123", 101, "testuser", true, "age"},
		{"username too short", "password123", 25, "ab", true, "username"},
		{"username too long", "password123", 25, "thisusernameiswaytoolong", true, "username"},
		{"all valid", "password123", 25, "testuser", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TestStruct{
				Email:    "test@example.com",
				Password: tt.password,
				Age:      tt.age,
				Username: tt.username,
			}

			err := validator.Validate(s)
			if tt.shouldError && err == nil {
				t.Error("expected validation error")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if tt.shouldError && err != nil {
				validationErrs, ok := err.(ValidationErrors)
				if !ok {
					t.Fatalf("expected ValidationErrors, got %T", err)
				}
				if _, exists := validationErrs[tt.errorField]; !exists {
					t.Errorf("expected validation error for field '%s'", tt.errorField)
				}
			}
		})
	}
}

func TestDefaultValidator_Validate_OptionalFields(t *testing.T) {
	validator := DefaultValidator()

	tests := []struct {
		name        string
		data        *OptionalFieldsStruct
		shouldError bool
	}{
		{
			name: "required field present, optional field empty",
			data: &OptionalFieldsStruct{
				Name:  "John",
				Email: "",
			},
			shouldError: false,
		},
		{
			name: "required field present, optional field valid",
			data: &OptionalFieldsStruct{
				Name:  "John",
				Email: "john@example.com",
			},
			shouldError: false,
		},
		{
			name: "required field missing",
			data: &OptionalFieldsStruct{
				Name:  "",
				Email: "john@example.com",
			},
			shouldError: true,
		},
		{
			name: "optional field invalid format",
			data: &OptionalFieldsStruct{
				Name:  "John",
				Email: "invalid-email",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
			if tt.shouldError && err == nil {
				t.Error("expected validation error")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		"email": "required",
		"age":   "min",
	}

	result := errs.Error()
	expected := "validation error"

	if result != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, result)
	}
}

func TestValidationErrors_MapStructure(t *testing.T) {
	validator := DefaultValidator()

	invalid := &TestStruct{
		Email:    "invalid",
		Password: "short",
		Age:      10,
		Username: "ab",
	}

	err := validator.Validate(invalid)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	validationErrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	// Check that errors are keyed by JSON field names
	if len(validationErrs) == 0 {
		t.Error("expected non-empty validation errors map")
	}

	// Verify each error has a proper tag
	for field, tag := range validationErrs {
		if tag == "" {
			t.Errorf("field '%s' has empty error tag", field)
		}
	}
}

func TestParseJSONTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected string
	}{
		{"simple tag", "email", "email"},
		{"tag with omitempty", "email,omitempty", "email"},
		{"tag with multiple options", "email,omitempty,string", "email"},
		{"tag with dash", "-", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseJSONTag(tt.tag)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

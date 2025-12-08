package errs

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewApiError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		msg            string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "bad request error",
			status:         http.StatusBadRequest,
			msg:            "invalid input",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid input",
		},
		{
			name:           "not found error",
			status:         http.StatusNotFound,
			msg:            "resource not found",
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "resource not found",
		},
		{
			name:           "internal server error",
			status:         http.StatusInternalServerError,
			msg:            "something went wrong",
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewApiError(tt.status, tt.msg)

			if err.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.Log != "" {
				t.Errorf("expected empty Log, got '%s'", err.Log)
			}
		})
	}
}

func TestApiError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError ApiError
		expected string
	}{
		{
			name:     "simple error message",
			apiError: NewApiError(http.StatusBadRequest, "bad request"),
			expected: "bad request",
		},
		{
			name:     "error with special characters",
			apiError: NewApiError(http.StatusNotFound, "user: not found!"),
			expected: "user: not found!",
		},
		{
			name:     "empty message",
			apiError: NewApiError(http.StatusInternalServerError, ""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiError.Error()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestApiError_Map(t *testing.T) {
	tests := []struct {
		name          string
		apiError      ApiError
		expectedError string
	}{
		{
			name:          "simple error",
			apiError:      NewApiError(http.StatusBadRequest, "validation failed"),
			expectedError: "validation failed",
		},
		{
			name:          "another error",
			apiError:      NewApiError(http.StatusUnauthorized, "unauthorized access"),
			expectedError: "unauthorized access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.apiError.Map()

			if len(m) != 1 {
				t.Errorf("expected map with 1 key, got %d", len(m))
			}

			if m["error"] != tt.expectedError {
				t.Errorf("expected error '%s', got '%s'", tt.expectedError, m["error"])
			}
		})
	}
}

func TestInternalServerError(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		expectedMsg string
		expectedLog string
	}{
		{
			name:        "database error",
			inputError:  errors.New("connection timeout"),
			expectedMsg: "internal server error",
			expectedLog: "connection timeout",
		},
		{
			name:        "generic error",
			inputError:  errors.New("unexpected error"),
			expectedMsg: "internal server error",
			expectedLog: "unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InternalServerError(tt.inputError)

			if err.StatusCode != http.StatusInternalServerError {
				t.Errorf("expected status %d, got %d", http.StatusInternalServerError, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.Log != tt.expectedLog {
				t.Errorf("expected log '%s', got '%s'", tt.expectedLog, err.Log)
			}
		})
	}
}

func TestUnauthorized(t *testing.T) {
	tests := []struct {
		name        string
		reason      string
		expectedMsg string
		expectedLog string
	}{
		{
			name:        "invalid token",
			reason:      "token expired",
			expectedMsg: "unauthorized",
			expectedLog: "token expired",
		},
		{
			name:        "missing credentials",
			reason:      "no credentials provided",
			expectedMsg: "unauthorized",
			expectedLog: "no credentials provided",
		},
		{
			name:        "empty reason",
			reason:      "",
			expectedMsg: "unauthorized",
			expectedLog: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unauthorized(tt.reason)

			if err.StatusCode != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.Log != tt.expectedLog {
				t.Errorf("expected log '%s', got '%s'", tt.expectedLog, err.Log)
			}
		})
	}
}

func TestForbidden(t *testing.T) {
	tests := []struct {
		name        string
		reason      string
		expectedMsg string
		expectedLog string
	}{
		{
			name:        "insufficient permissions",
			reason:      "user lacks required role",
			expectedMsg: "forbidden",
			expectedLog: "user lacks required role",
		},
		{
			name:        "access denied",
			reason:      "resource access denied",
			expectedMsg: "forbidden",
			expectedLog: "resource access denied",
		},
		{
			name:        "empty reason",
			reason:      "",
			expectedMsg: "forbidden",
			expectedLog: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Forbidden(tt.reason)

			if err.StatusCode != http.StatusForbidden {
				t.Errorf("expected status %d, got %d", http.StatusForbidden, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.Log != tt.expectedLog {
				t.Errorf("expected log '%s', got '%s'", tt.expectedLog, err.Log)
			}
		})
	}
}

func TestInvalidJson(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		expectedMsg string
		expectedLog string
	}{
		{
			name:        "json parse error",
			inputError:  errors.New("unexpected end of JSON input"),
			expectedMsg: "invalid json",
			expectedLog: "unexpected end of JSON input",
		},
		{
			name:        "syntax error",
			inputError:  errors.New("invalid character"),
			expectedMsg: "invalid json",
			expectedLog: "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidJson(tt.inputError)

			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.Log != tt.expectedLog {
				t.Errorf("expected log '%s', got '%s'", tt.expectedLog, err.Log)
			}
		})
	}
}

func TestInvalidFormData(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		expectedMsg string
		expectedLog string
	}{
		{
			name:        "missing field",
			inputError:  errors.New("required field missing"),
			expectedMsg: "invalid form data",
			expectedLog: "required field missing",
		},
		{
			name:        "validation error",
			inputError:  errors.New("email format invalid"),
			expectedMsg: "invalid form data",
			expectedLog: "email format invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidFormData(tt.inputError)
			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.Log != tt.expectedLog {
				t.Errorf("expected log '%s', got '%s'", tt.expectedLog, err.Log)
			}
		})
	}
}

func TestInvalidPathParam(t *testing.T) {
	tests := []struct {
		name        string
		param       string
		expectedMsg string
	}{
		{
			name:        "id param",
			param:       "id",
			expectedMsg: "invalid path param: id",
		},
		{
			name:        "userId param",
			param:       "userId",
			expectedMsg: "invalid path param: userId",
		},
		{
			name:        "empty param",
			param:       "",
			expectedMsg: "invalid path param: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidPathParam(tt.param)

			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
		})
	}
}

func TestInvalidQueryParam(t *testing.T) {
	tests := []struct {
		name        string
		param       string
		expectedMsg string
	}{
		{
			name:        "limit param",
			param:       "limit",
			expectedMsg: "invalid query param: limit",
		},
		{
			name:        "offset param",
			param:       "offset",
			expectedMsg: "invalid query param: offset",
		},
		{
			name:        "empty param",
			param:       "",
			expectedMsg: "invalid query param: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidQueryParam(tt.param)

			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		name        string
		inputError  string
		expectedMsg string
		expectedLog string
	}{
		{
			name:        "user not found",
			inputError:  "user not found in database",
			expectedMsg: "not found",
			expectedLog: "user not found in database",
		},
		{
			name:        "resource not found",
			inputError:  "resource does not exist",
			expectedMsg: "not found",
			expectedLog: "resource does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NotFound(tt.inputError)

			if err.StatusCode != http.StatusNotFound {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.Log != tt.expectedLog {
				t.Errorf("expected log '%s', got '%s'", tt.expectedLog, err.Log)
			}
		})
	}
}

func TestObjectNotFound(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		resource    string
		expectedMsg string
	}{
		{
			name:        "user not found",
			id:          "123",
			resource:    "user",
			expectedMsg: "user with ID 123 not found",
		},
		{
			name:        "resource not found",
			id:          "456",
			resource:    "product",
			expectedMsg: "product with ID 456 not found",
		},
		{
			name:        "empty id and resource",
			id:          "",
			resource:    "",
			expectedMsg: " with ID  not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ObjectNotFound(tt.id, tt.resource)

			if err.StatusCode != http.StatusNotFound {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.Log != "" {
				t.Errorf("expected empty Log, got '%s'", err.Log)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		data           any
		expectedStatus int
		expectedMsg    string
		expectedLog    string
		expectedData   any
	}{
		{
			name: "validation error with map data",
			err:  errors.New("field validation failed"),
			data: map[string]string{
				"email": "required",
				"age":   "min",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "validation error",
			expectedLog:    "field validation failed",
			expectedData: map[string]string{
				"email": "required",
				"age":   "min",
			},
		},
		{
			name: "validation error with struct data",
			err:  errors.New("validation failed"),
			data: struct {
				Field   string
				Message string
			}{
				Field:   "username",
				Message: "too short",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "validation error",
			expectedLog:    "validation failed",
		},
		{
			name:           "validation error with nil data",
			err:            errors.New("some validation error"),
			data:           nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "validation error",
			expectedLog:    "some validation error",
			expectedData:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := ValidationError(tt.err, tt.data)

			if apiErr.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, apiErr.StatusCode)
			}
			if apiErr.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, apiErr.Msg)
			}
			if apiErr.Log != tt.expectedLog {
				t.Errorf("expected log '%s', got '%s'", tt.expectedLog, apiErr.Log)
			}
			if tt.expectedData != nil && apiErr.Data == nil {
				t.Error("expected Data to be set, got nil")
			}

			// Test Map() method
			resultMap := apiErr.Map()
			if resultMap["error"] != tt.expectedMsg {
				t.Errorf("expected map error '%s', got '%v'", tt.expectedMsg, resultMap["error"])
			}
			if tt.data != nil {
				if _, hasData := resultMap["data"]; !hasData {
					t.Error("expected map to have 'data' key")
				}
			}
		})
	}
}

func TestInvalidBody(t *testing.T) {
	tests := []struct {
		name           string
		validationErrs any
		expectedStatus int
		expectedMsg    string
		expectedData   any
	}{
		{
			name: "validation errors as map",
			validationErrs: map[string]string{
				"email":    "required",
				"password": "min",
				"age":      "min",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request body",
			expectedData: map[string]string{
				"email":    "required",
				"password": "min",
				"age":      "min",
			},
		},
		{
			name: "validation errors as custom type",
			validationErrs: map[string]interface{}{
				"field1": map[string]string{
					"error": "invalid format",
					"value": "abc123",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request body",
		},
		{
			name:           "empty validation errors",
			validationErrs: map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request body",
			expectedData:   map[string]string{},
		},
		{
			name:           "nil validation errors",
			validationErrs: nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request body",
			expectedData:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := InvalidBody(tt.validationErrs)

			if apiErr.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, apiErr.StatusCode)
			}
			if apiErr.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, apiErr.Msg)
			}
			if apiErr.Data == nil && tt.validationErrs != nil {
				t.Error("expected Data to be set, got nil")
			}
			if apiErr.Log != "" {
				t.Errorf("expected empty Log, got '%s'", apiErr.Log)
			}

			// Test Map() method includes data
			resultMap := apiErr.Map()
			if resultMap["error"] != tt.expectedMsg {
				t.Errorf("expected map error '%s', got '%v'", tt.expectedMsg, resultMap["error"])
			}
			if tt.validationErrs != nil {
				if _, hasData := resultMap["data"]; !hasData {
					t.Error("expected map to have 'data' key when validationErrs is not nil")
				}
			}
		})
	}
}

func TestInvalidBody_MapStructure(t *testing.T) {
	validationErrs := map[string]string{
		"email":    "email",
		"password": "min",
		"username": "required",
	}

	apiErr := InvalidBody(validationErrs)
	resultMap := apiErr.Map()

	if len(resultMap) != 2 {
		t.Errorf("expected map with 2 keys, got %d", len(resultMap))
	}

	errorMsg, ok := resultMap["error"].(string)
	if !ok {
		t.Fatal("expected 'error' to be a string")
	}
	if errorMsg != "invalid request body" {
		t.Errorf("expected error 'invalid request body', got '%s'", errorMsg)
	}

	data, ok := resultMap["data"].(map[string]string)
	if !ok {
		t.Fatal("expected 'data' to be a map[string]string")
	}
	if len(data) != 3 {
		t.Errorf("expected data with 3 fields, got %d", len(data))
	}

	expectedFields := []string{"email", "password", "username"}
	for _, field := range expectedFields {
		if _, exists := data[field]; !exists {
			t.Errorf("expected data to contain field '%s'", field)
		}
	}
}

func TestValidationError_vs_InvalidBody(t *testing.T) {
	validationData := map[string]string{"field": "error"}

	valErr := ValidationError(errors.New("validation failed"), validationData)
	bodyErr := InvalidBody(validationData)

	if valErr.StatusCode != bodyErr.StatusCode {
		t.Error("both errors should have same status code")
	}

	if valErr.Msg == bodyErr.Msg {
		t.Error("errors should have different messages")
	}

	if valErr.Log == "" {
		t.Error("validation error should have log set")
	}
	if bodyErr.Log != "" {
		t.Error("invalid body error should have empty log")
	}

	valMap := valErr.Map()
	bodyMap := bodyErr.Map()

	if _, hasData := valMap["data"]; !hasData {
		t.Error("validation error map should have 'data' key")
	}
	if _, hasData := bodyMap["data"]; !hasData {
		t.Error("invalid body error map should have 'data' key")
	}
}

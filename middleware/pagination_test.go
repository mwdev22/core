package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mwdev22/rest/cctx"
)

func TestPagination(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		defaultLimit   int
		maxLimit       int
		expectedLimit  int
		expectedOffset int
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "default values - no query params",
			query:          "",
			defaultLimit:   10,
			maxLimit:       100,
			expectedLimit:  10,
			expectedOffset: 0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "page 1",
			query:          "?page=1",
			defaultLimit:   10,
			maxLimit:       100,
			expectedLimit:  10,
			expectedOffset: 0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "page 2",
			query:          "?page=2",
			defaultLimit:   10,
			maxLimit:       100,
			expectedLimit:  10,
			expectedOffset: 10,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "page 3 with custom limit",
			query:          "?page=3&limit=20",
			defaultLimit:   10,
			maxLimit:       100,
			expectedLimit:  20,
			expectedOffset: 40,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom limit only",
			query:          "?limit=25",
			defaultLimit:   10,
			maxLimit:       100,
			expectedLimit:  25,
			expectedOffset: 0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "page 5 with limit 15",
			query:          "?page=5&limit=15",
			defaultLimit:   10,
			maxLimit:       100,
			expectedLimit:  15,
			expectedOffset: 60,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid page - not a number",
			query:          "?page=abc",
			defaultLimit:   10,
			maxLimit:       100,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid limit - not a number",
			query:          "?limit=xyz",
			defaultLimit:   10,
			maxLimit:       100,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "page 0 - edge case",
			query:          "?page=0",
			defaultLimit:   10,
			maxLimit:       100,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "negative page",
			query:          "?page=-1",
			defaultLimit:   10,
			maxLimit:       100,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "negative limit",
			query:          "?limit=-5",
			defaultLimit:   10,
			maxLimit:       100,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "zero limit",
			query:          "?limit=0",
			defaultLimit:   10,
			maxLimit:       100,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "limit exceeds maximum",
			query:          "?limit=101",
			defaultLimit:   10,
			maxLimit:       100,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "limit at maximum (100)",
			query:          "?limit=100",
			defaultLimit:   10,
			maxLimit:       100,
			expectedLimit:  100,
			expectedOffset: 0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "limit at minimum (1)",
			query:          "?limit=1",
			defaultLimit:   10,
			maxLimit:       100,
			expectedLimit:  1,
			expectedOffset: 0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom maxLimit - 50",
			query:          "?limit=50",
			defaultLimit:   10,
			maxLimit:       50,
			expectedLimit:  50,
			expectedOffset: 0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "exceeds custom maxLimit",
			query:          "?limit=51",
			defaultLimit:   10,
			maxLimit:       50,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Pagination(tt.defaultLimit, tt.maxLimit)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				limit, _ := r.Context().Value(cctx.Limit).(int)
				offset, _ := r.Context().Value(cctx.Offset).(int)

				if !tt.expectError {
					if limit != tt.expectedLimit {
						t.Errorf("expected limit %d, got %d", tt.expectedLimit, limit)
					}
					if offset != tt.expectedOffset {
						t.Errorf("expected offset %d, got %d", tt.expectedOffset, offset)
					}
				}

				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectError {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if _, hasError := response["error"]; !hasError {
					t.Error("expected error response to have 'error' field")
				}
			}
		})
	}
}

func TestPagination_ContextValues(t *testing.T) {
	middleware := Pagination(10, 100)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.Context().Value(cctx.Limit)
		offset := r.Context().Value(cctx.Offset)

		if limit == nil {
			t.Error("expected limit in context, got nil")
		}
		if offset == nil {
			t.Error("expected offset in context, got nil")
		}

		limitInt, ok := limit.(int)
		if !ok {
			t.Errorf("expected limit to be int, got %T", limit)
		}
		if limitInt != 50 {
			t.Errorf("expected limit 50, got %d", limitInt)
		}

		offsetInt, ok := offset.(int)
		if !ok {
			t.Errorf("expected offset to be int, got %T", offset)
		}
		if offsetInt != 100 {
			t.Errorf("expected offset 100, got %d", offsetInt)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test?page=3&limit=50", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestPagination_DefaultLimitValues(t *testing.T) {
	tests := []struct {
		name         string
		defaultLimit int
		maxLimit     int
		query        string
		expected     int
	}{
		{"default 10", 10, 100, "", 10},
		{"default 20", 20, 100, "", 20},
		{"default 50", 50, 100, "", 50},
		{"override default", 10, 100, "?limit=30", 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Pagination(tt.defaultLimit, tt.maxLimit)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				limit := r.Context().Value(cctx.Limit).(int)
				if limit != tt.expected {
					t.Errorf("expected limit %d, got %d", tt.expected, limit)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
		})
	}
}

func TestPagination_CustomMaxLimit(t *testing.T) {
	tests := []struct {
		name          string
		maxLimit      int
		query         string
		expectError   bool
		expectedLimit int
	}{
		{"maxLimit 50 - within limit", 50, "?limit=40", false, 40},
		{"maxLimit 50 - at limit", 50, "?limit=50", false, 50},
		{"maxLimit 50 - exceeds", 50, "?limit=51", true, 0},
		{"maxLimit 200 - within limit", 200, "?limit=150", false, 150},
		{"maxLimit 200 - exceeds", 200, "?limit=201", true, 0},
		{"maxLimit 25 - within limit", 25, "?limit=20", false, 20},
		{"maxLimit 25 - exceeds", 25, "?limit=30", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Pagination(10, tt.maxLimit)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectError {
					t.Error("handler should not be called on error")
					return
				}
				limit := r.Context().Value(cctx.Limit).(int)
				if limit != tt.expectedLimit {
					t.Errorf("expected limit %d, got %d", tt.expectedLimit, limit)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.expectError && w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 error, got %d", w.Code)
			}
			if !tt.expectError && w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", w.Code)
			}
		})
	}
}

func TestPagination_OffsetCalculation(t *testing.T) {
	tests := []struct {
		page           int
		limit          int
		expectedOffset int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 10, 20},
		{1, 20, 0},
		{2, 20, 20},
		{5, 15, 60},
		{10, 100, 900},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			middleware := Pagination(10, 100)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				offset := r.Context().Value(cctx.Offset).(int)
				if offset != tt.expectedOffset {
					t.Errorf("page=%d, limit=%d: expected offset %d, got %d",
						tt.page, tt.limit, tt.expectedOffset, offset)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet,
				"/test?page="+strconv.Itoa(tt.page)+"&limit="+strconv.Itoa(tt.limit), nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
		})
	}
}

func TestPagination_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedError string
	}{
		{
			name:          "invalid page",
			query:         "?page=notanumber",
			expectedError: "page must be a number",
		},
		{
			name:          "invalid limit",
			query:         "?limit=notanumber",
			expectedError: "limit must be a number",
		},
		{
			name:          "page too small",
			query:         "?page=0",
			expectedError: "page must be at least 1",
		},
		{
			name:          "negative page",
			query:         "?page=-5",
			expectedError: "page must be at least 1",
		},
		{
			name:          "limit too small",
			query:         "?limit=0",
			expectedError: "limit must be at least 1",
		},
		{
			name:          "negative limit",
			query:         "?limit=-10",
			expectedError: "limit must be at least 1",
		},
		{
			name:          "limit too large",
			query:         "?limit=101",
			expectedError: "limit must not exceed 100",
		},
		{
			name:          "limit way too large",
			query:         "?limit=999999",
			expectedError: "limit must not exceed 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Pagination(10, 100)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called on error")
			}))

			req := httptest.NewRequest(http.MethodGet, "/test"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}

			var response map[string]string
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response["error"] != tt.expectedError {
				t.Errorf("expected error '%s', got '%s'", tt.expectedError, response["error"])
			}
		})
	}
}

func TestPagination_BoundaryValues(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedLimit int
		shouldError   bool
	}{
		{"minimum valid limit", "?limit=1", 1, false},
		{"maximum valid limit", "?limit=100", 100, false},
		{"below minimum", "?limit=0", 0, true},
		{"above maximum", "?limit=101", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Pagination(10, 100)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.shouldError {
					t.Error("handler should not be called on error")
					return
				}
				limit := r.Context().Value(cctx.Limit).(int)
				if limit != tt.expectedLimit {
					t.Errorf("expected limit %d, got %d", tt.expectedLimit, limit)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.shouldError && w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 error, got %d", w.Code)
			}
			if !tt.shouldError && w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", w.Code)
			}
		})
	}
}

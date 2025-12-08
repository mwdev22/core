package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mwdev22/rest/cctx"
	"github.com/mwdev22/rest/utils/errs"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name           string
		handler        HandlerWithErr
		expectedStatus int
		expectedBody   string
		checkJSON      bool
		expectedError  string
	}{
		{
			name: "ok - no error",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "error - invalid json",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errs.InvalidJson(errors.New("unexpected token"))
			},
			expectedStatus: http.StatusBadRequest,
			checkJSON:      true,
			expectedError:  "invalid json",
		},
		{
			name: "error - invalid path param",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errs.InvalidPathParam("id")
			},
			expectedStatus: http.StatusBadRequest,
			checkJSON:      true,
			expectedError:  "invalid path param: id",
		},
		{
			name: "error - invalid query param",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errs.InvalidQueryParam("id")
			},
			expectedStatus: http.StatusBadRequest,
			checkJSON:      true,
			expectedError:  "invalid query param: id",
		},
		{
			name: "error - not found",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errs.NotFound("object not found")
			},
			expectedStatus: http.StatusNotFound,
			checkJSON:      true,
			expectedError:  "not found",
		},
		{
			name: "generic error",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errors.New("something went wrong")
			},
			expectedStatus: http.StatusInternalServerError,
			checkJSON:      true,
			expectedError:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Wrap(tt.handler)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkJSON {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("expected error message '%s', got '%s'", tt.expectedError, response["error"])
				}
			} else if tt.expectedBody != "" {
				if w.Body.String() != tt.expectedBody {
					t.Errorf("expected body '%s', got '%s'", tt.expectedBody, w.Body.String())
				}
			}
		})
	}
}

func TestRealIP(t *testing.T) {
	tests := []struct {
		name       string
		setupReq   func(*http.Request)
		expectedIP string
	}{
		{
			name: "X-Forwarded-For header",
			setupReq: func(req *http.Request) {
				req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
			},
			expectedIP: "203.0.113.1",
		},
		{
			name: "X-Real-IP header",
			setupReq: func(req *http.Request) {
				req.Header.Set("X-Real-IP", "198.51.100.42")
			},
			expectedIP: "198.51.100.42",
		},
		{
			name: "RemoteAddr fallback",
			setupReq: func(req *http.Request) {
				req.RemoteAddr = "192.0.2.1:12345"
			},
			expectedIP: "192.0.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ip, ok := r.Context().Value(cctx.RealIpKey).(string)
				if !ok {
					t.Error("RealIpKey not found in context")
					return
				}
				if ip != tt.expectedIP {
					t.Errorf("expected IP %s, got %s", tt.expectedIP, ip)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			tt.setupReq(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestInternal(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		expectedStatus int
		expectedError  string
		shouldAllow    bool
	}{
		{
			name:           "allow 192.168 network",
			ip:             "192.168.1.100",
			expectedStatus: http.StatusOK,
			shouldAllow:    true,
		},

		{
			name:           "block public IP",
			ip:             "203.0.113.1",
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			shouldAllow:    false,
		},
		{
			name:           "block loopback",
			ip:             "127.0.0.1",
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			shouldAllow:    false,
		},
		{
			name:           "block empty IP",
			ip:             "",
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			shouldAllow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Internal(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("allowed"))
			}))

			req := httptest.NewRequest(http.MethodGet, "/internal", nil)
			req = req.WithContext(context.WithValue(req.Context(), cctx.RealIpKey, tt.ip))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d for IP %s, got %d", tt.expectedStatus, tt.ip, w.Code)
			}

			if tt.shouldAllow {
				if w.Body.String() != "allowed" {
					t.Errorf("expected body 'allowed', got %s", w.Body.String())
				}
			} else {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("expected error '%s', got '%s'", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestLogger(t *testing.T) {
	loggerMiddleware := Logger(nil)
	handler := loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("logged"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "logged" {
		t.Errorf("expected body 'logged', got %s", w.Body.String())
	}
}

func TestRecoverer(t *testing.T) {
	handler := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 after panic, got %d", w.Code)
	}
}

func TestColorMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected string
	}{
		{"GET", colorBlue + "GET" + colorReset},
		{"POST", colorGreen + "POST" + colorReset},
		{"PUT", colorYellow + "PUT" + colorReset},
		{"DELETE", colorRed + "DELETE" + colorReset},
		{"PATCH", colorCyan + "PATCH" + colorReset},
		{"OPTIONS", colorCyan + "OPTIONS" + colorReset},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := colorMethod(tt.method)
			if result != tt.expected {
				t.Errorf("colorMethod(%s) = %q, want %q", tt.method, result, tt.expected)
			}
		})
	}
}

func TestColorStatus(t *testing.T) {
	tests := []struct {
		status   int
		expected string
	}{
		{200, colorGreen + "200" + colorReset},
		{299, colorGreen + "299" + colorReset},
		{301, colorYellow + "301" + colorReset},
		{404, colorRed + "404" + colorReset},
		{500, colorRed + "500" + colorReset},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.status), func(t *testing.T) {
			result := colorStatus(tt.status)
			if result != tt.expected {
				t.Errorf("colorStatus(%d) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/mwdev22/logging"
	"github.com/mwdev22/rest/cctx"
	"github.com/mwdev22/rest/jsonutil"
	"github.com/mwdev22/rest/utils/errs"
)

type ApiHandler func(w http.ResponseWriter, r *http.Request) error
type MiddlewareToChain func(next http.Handler) http.Handler

// allows handlers to return errors
func Wrap(final ApiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := final(w, r); err != nil {
			var e errs.ApiError
			l := logging.FromContext(r.Context())
			if errors.As(err, &e) {
				jsonutil.Write(w, e.StatusCode, e.Map())
				l.Printf("%sAPI ERROR%s: %s", colorRed, colorReset, e.Log)
			} else {
				jsonutil.Write(w, http.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
				l.Printf("%sUNKNOWN ERROR%s: %s", colorRed, colorReset, err.Error())
			}
		}
	}
}

func Pagination(defaultLimit int, maxLimit int) MiddlewareToChain {
	const minLimit = 1
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				pageQ := r.URL.Query().Get("page")
				limitQ := r.URL.Query().Get("limit")
				limit, offset := defaultLimit, 0

				if limitQ != "" {
					l, err := strconv.Atoi(limitQ)
					if err != nil {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": "limit must be a number",
						})
						return
					}
					if l < minLimit {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": "limit must be at least 1",
						})
						return
					}
					if l > maxLimit {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": "limit must not exceed 100",
						})
						return
					}
					limit = l
				}

				if pageQ != "" {
					p, err := strconv.Atoi(pageQ)
					if err != nil {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": "page must be a number",
						})
						return
					}
					if p < 1 {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": "page must be at least 1",
						})
						return
					}
					offset = (p - 1) * limit
				}

				ctx := context.WithValue(r.Context(), cctx.Offset, offset)
				ctx = context.WithValue(ctx, cctx.Limit, limit)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
	}
}

func Logger(l logging.Logger) MiddlewareToChain {
	if l == nil {
		l = logging.DefaultLogger()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			before := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			ctx := logging.ToContext(r.Context(), l)
			defer func() {
				duration := time.Since(before)
				l.Printf("[%s] %s %s %.2fms",
					colorMethod(r.Method),
					r.RequestURI,
					colorStatus(ww.Status()),
					float64(duration.Microseconds())/1000.0)
			}()
			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}
func RateLimit(limit int, windowLength time.Duration) MiddlewareToChain {
	return func(next http.Handler) http.Handler {
		return httprate.LimitByRealIP(limit, windowLength)(next)
	}
}

func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := func(r *http.Request) string {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				return strings.TrimSpace(strings.Split(xff, ",")[0])
			}
			if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
				return xrip
			}
			host, _, _ := strings.Cut(r.RemoteAddr, ":")
			return host
		}(r)

		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), cctx.RealIpKey, ip)),
		)
	})
}

func Internal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _ := r.Context().Value(cctx.RealIpKey).(string)
		if ip != "" {
			l := logging.FromContext(r.Context())
			l.Printf("internal route ‑ caller IP: %s", ip)
		}

		if !strings.HasPrefix(ip, "192.168.") {
			_ = jsonutil.Write(w, http.StatusForbidden, map[string]string{
				"error": "forbidden",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

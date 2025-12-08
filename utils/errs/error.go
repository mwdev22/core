package errs

import (
	"fmt"
	"net/http"

	"github.com/mwdev22/rest/cctx"
)

type ApiError struct {
	StatusCode int
	Msg        string
	Log        string
	Data       any
}

func (e ApiError) Error() string {
	return e.Msg
}

func (e ApiError) Map() map[string]any {
	if e.Data != nil {
		return map[string]any{
			"error": e.Error(),
			"data":  e.Data,
		}
	}
	return map[string]any{
		"error": e.Error(),
	}
}

func NewApiError(status int, msg string) ApiError {
	return ApiError{
		StatusCode: status,
		Msg:        msg,
	}
}

func InternalServerError(err error) ApiError {
	return ApiError{
		StatusCode: http.StatusInternalServerError,
		Msg:        "internal server error",
		Log:        err.Error(),
	}
}

func Unauthorized(reason string) ApiError {
	return ApiError{
		StatusCode: http.StatusUnauthorized,
		Msg:        "unauthorized",
		Log:        reason,
	}
}

func Forbidden(reason string) ApiError {
	return ApiError{
		StatusCode: http.StatusForbidden,
		Msg:        "forbidden",
		Log:        reason,
	}
}

func InvalidJson(err error) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        "invalid json",
		Log:        err.Error(),
	}
}

func InvalidFormData(err error) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        "invalid form data",
		Log:        err.Error(),
	}
}

func InvalidPathParam(param string) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        fmt.Sprintf("invalid path param: %s", param),
	}
}

func InvalidQueryParam(param string) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        fmt.Sprintf("invalid query param: %s", param),
	}
}

func MissingContextValue(key cctx.ContextKey) ApiError {
	return ApiError{
		StatusCode: http.StatusInternalServerError,
		Msg:        "internal server error",
		Log:        fmt.Sprintf("missing context value for key: %s", key),
	}
}

func NotFound(reason string) ApiError {
	return ApiError{
		StatusCode: http.StatusNotFound,
		Msg:        "not found",
		Log:        reason,
	}
}

func ObjectNotFound(id string, name string) ApiError {
	return ApiError{
		StatusCode: http.StatusNotFound,
		Msg:        fmt.Sprintf("%s with ID %s not found", name, id),
		Log:        "",
	}
}

func ValidationError(err error, data any) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        "validation error",
		Log:        err.Error(),
		Data:       data,
	}
}

func InvalidBody(validationErrs any) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        "invalid request body",
		Data:       validationErrs,
	}
}

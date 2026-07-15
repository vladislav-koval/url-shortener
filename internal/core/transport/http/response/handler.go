package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/core/errors"
	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"go.uber.org/zap"
)

type HTTPResponseHandler struct {
	log *logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(log *logger.Logger, rw http.ResponseWriter) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  rw,
	}
}

func (h *HTTPResponseHandler) JSONResponse(responseBody any, statusCode int) {
	h.rw.Header().Set("Content-Type", "application/json")

	h.rw.WriteHeader(statusCode)

	if err := json.NewEncoder(h.rw).Encode(responseBody); err != nil {
		h.log.Error("failed to encode response body", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) HTMLResponse(html []byte) {
	h.rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	h.rw.WriteHeader(http.StatusOK)
	if _, err := h.rw.Write(html); err != nil {
		h.log.Error("failed to write HTML HTTP response", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) NoContentResponse() {
	h.rw.Header().Set("Content-Type", "application/json")

	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *HTTPResponseHandler) RedirectResponse(r *http.Request, url string) {
	http.Redirect(h.rw, r, url, http.StatusFound)
}

func (h *HTTPResponseHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode int
		code       string
		logFunc    func(string, ...zap.Field)
	)

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		statusCode, code, logFunc = http.StatusNotFound, "not_found", h.log.Debug

	case errors.Is(err, apperrors.ErrInvalidArgument):
		statusCode, code, logFunc = http.StatusBadRequest, "invalid_argument", h.log.Warn

	case errors.Is(err, apperrors.ErrConflict):
		statusCode, code, logFunc = http.StatusConflict, "conflict", h.log.Warn

	default:
		statusCode, code, logFunc = http.StatusInternalServerError, "internal_error", h.log.Error
	}

	logFunc(msg, zap.Error(err))

	resp := ErrorResponse{Code: code, Message: msg}

	var fieldErrs apperrors.ValidationErrors
	if errors.As(err, &fieldErrs) {
		resp.Details = make([]FieldError, 0, len(fieldErrs))
		for _, fe := range fieldErrs {
			resp.Details = append(resp.Details, FieldError{Field: fe.Field, Rule: fe.Rule, Param: fe.Param})
		}
	}

	h.JSONResponse(resp, statusCode)
}

func (h *HTTPResponseHandler) PanicResponse(p any, msg string) {
	err := fmt.Errorf("unexpected panic: %v", p)

	h.log.Error(msg, zap.Error(err))

	h.JSONResponse(ErrorResponse{Code: "internal_error", Message: msg}, http.StatusInternalServerError)
}

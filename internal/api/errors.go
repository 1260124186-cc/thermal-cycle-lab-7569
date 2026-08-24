package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"thermal-cycle-lab/internal/store"
)

func writeDomainError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "finalize") {
		writeError(w, http.StatusInternalServerError, "finalization processing failed")
		return
	}
	switch {
	case errors.Is(err, context.Canceled):
		writeError(w, 499, "request was canceled")
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "request deadline exceeded")
	case store.IsNotFound(err):
		writeError(w, http.StatusNotFound, err.Error())
	case store.IsConflict(err):
		writeError(w, http.StatusConflict, err.Error())
	case isValidationError(err):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}
func isValidationError(err error) bool {
	text := strings.ToLower(err.Error())
	markers := []string{"required", "invalid", "cannot", "must", "unsafe", "outside", "earlier", "advance", "tolerance", "available", "accept"}
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ryoshimaru/MindMap/internal/store"
)

type DataEnvelope struct {
	Data any `json:"data"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(DataEnvelope{Data: data})
}

func WriteError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{
		Error: ErrorBody{Code: code, Message: message},
	})
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return false
	}
	return true
}

func WriteStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrValidation):
		WriteError(w, http.StatusBadRequest, "validation_error", "request data is invalid")
	case errors.Is(err, store.ErrConflict):
		WriteError(w, http.StatusConflict, "conflict", "resource already exists")
	case errors.Is(err, store.ErrInvalidAuth):
		WriteError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
	case errors.Is(err, store.ErrInvalidState):
		WriteError(w, http.StatusConflict, "invalid_request_state", "the request is not in a state that allows this operation")
	case errors.Is(err, store.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
	case errors.Is(err, store.ErrNotFound):
		WriteError(w, http.StatusNotFound, "not_found", "resource was not found")
	default:
		WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

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
		WriteError(w, http.StatusBadRequest, "invalid_json", "Некорректный формат данных")
		return false
	}
	return true
}

func WriteStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrValidation):
		WriteError(w, http.StatusBadRequest, "validation_error", "Проверьте заполненные данные")
	case errors.Is(err, store.ErrConflict):
		WriteError(w, http.StatusConflict, "conflict", "Такая запись уже существует")
	case errors.Is(err, store.ErrInvalidAuth):
		WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Неверная почта или пароль")
	case errors.Is(err, store.ErrInvalidAIKey):
		WriteError(w, http.StatusUnprocessableEntity, "invalid_ai_key", "AI-провайдер отклонил ключ. Проверьте ключ и повторите попытку")
	case errors.Is(err, store.ErrInvalidState):
		WriteError(w, http.StatusConflict, "invalid_request_state", "Действие сейчас недоступно: завершите предыдущие шаги или обязательную настройку")
	case errors.Is(err, store.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
	case errors.Is(err, store.ErrNotFound):
		WriteError(w, http.StatusNotFound, "not_found", "Данные не найдены")
	default:
		WriteError(w, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка сервера")
	}
}

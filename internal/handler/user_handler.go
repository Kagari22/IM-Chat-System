package handler

import (
	"net/http"

	"IM_Chat_System/internal/httpx"
	"IM_Chat_System/internal/service"
)

type UserHandler struct {
	messages *service.MessageService
}

func NewUserHandler(messages *service.MessageService) *UserHandler {
	return &UserHandler{messages: messages}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	user, err := h.messages.GetMe(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *UserHandler) Users(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	users, err := h.messages.ListUsers(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"users": users})
}

package handler

import (
	"net/http"
	"strconv"

	"IM_Chat_System/internal/httpx"
	"IM_Chat_System/internal/service"
)

type MediaHandler struct {
	messages *service.MessageService
}

func NewMediaHandler(messages *service.MessageService) *MediaHandler {
	return &MediaHandler{messages: messages}
}

func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	toUserID, _ := strconv.ParseInt(r.FormValue("to_user_id"), 10, 64)
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	message, err := h.messages.SaveMedia(r.Context(), claims.UserID, toUserID, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"message": message})
}

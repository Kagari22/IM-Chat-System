package handler

import (
	"net/http"
	"time"

	"IM_Chat_System/internal/httpx"
	"IM_Chat_System/internal/tokenblacklist"
)

type LogoutHandler struct {
	blacklist tokenblacklist.Store
}

func NewLogoutHandler(blacklist tokenblacklist.Store) *LogoutHandler {
	return &LogoutHandler{blacklist: blacklist}
}

func (h *LogoutHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	token, ok := TokenFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing token")
		return
	}
	if h.blacklist == nil {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	ttl := time.Until(time.Unix(claims.Expires, 0))
	if err := h.blacklist.Blacklist(r.Context(), token, ttl); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "logout failed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

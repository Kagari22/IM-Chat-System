package handler

import (
	"net/http"
	"strconv"

	"IM_Chat_System/internal/httpx"
	"IM_Chat_System/internal/search"
)

type SearchHandler struct {
	indexer search.Indexer
}

func NewSearchHandler(indexer search.Indexer) *SearchHandler {
	return &SearchHandler{indexer: indexer}
}

func (h *SearchHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	query := r.URL.Query().Get("q")
	peerID, _ := strconv.ParseInt(r.URL.Query().Get("peer_id"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	messages, err := h.indexer.SearchMessages(r.Context(), claims.UserID, query, peerID, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

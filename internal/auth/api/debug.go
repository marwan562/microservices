package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sapliy/fintech-ecosystem/internal/flow"
	flowDomain "github.com/sapliy/fintech-ecosystem/internal/flow/domain"
	"github.com/sapliy/fintech-ecosystem/pkg/jsonutil"
)

type DebugHandler struct {
	debugService *flow.DebugService
}

func NewDebugHandler(debugService *flow.DebugService) *DebugHandler {
	return &DebugHandler{debugService: debugService}
}

func (h *DebugHandler) StartDebugSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FlowID string `json:"flow_id"`
		ZoneID string `json:"zone_id"`
		Level  string `json:"level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	level := flowDomain.DebugLevel(req.Level)
	session, err := h.debugService.StartDebugSession(context.Background(), req.FlowID, req.ZoneID, level)
	if err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusCreated, session)
}

func (h *DebugHandler) GetDebugSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	session, err := h.debugService.GetDebugSession(sessionID)
	if err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}
	jsonutil.WriteJSON(w, http.StatusOK, session)
}

func (h *DebugHandler) WebSocketDebug(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		oneHourAgo := time.Now().Add(-1 * time.Hour)
		newEvents, err := h.debugService.GetDebugEvents(sessionID, &oneHourAgo)
		if err != nil {
			continue
		}
		for _, event := range newEvents {
			eventJSON, _ := json.Marshal(event)
			if err := conn.WriteMessage(websocket.TextMessage, eventJSON); err != nil {
				return
			}
		}
	}
}

func (h *DebugHandler) GetDebugEvents(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	var since *time.Time
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if timestamp, err := strconv.ParseInt(sinceStr, 10, 64); err == nil {
			t := time.Unix(timestamp/1000, 0)
			since = &t
		}
	}

	events, err := h.debugService.GetDebugEvents(sessionID, since)
	if err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, events)
}
func (h *DebugHandler) EndDebugSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if err := h.debugService.EndDebugSession(sessionID); err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ended"})
}

func (h *DebugHandler) ExecuteFlowWithDebug(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FlowID string                 `json:"flow_id"`
		Input  map[string]interface{} `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}
	// Simplified for migration
	jsonutil.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "executing"})
}

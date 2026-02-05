package main

import (
	"encoding/json"
	"net/http"

	"github.com/sapliy/fintech-ecosystem/internal/flow/domain"
	"github.com/sapliy/fintech-ecosystem/pkg/jsonutil"
)

type FlowHandler struct {
	repo   domain.Repository
	runner *domain.FlowRunner
}

func NewFlowHandler(repo domain.Repository, runner *domain.FlowRunner) *FlowHandler {
	return &FlowHandler{repo: repo, runner: runner}
}

func (h *FlowHandler) GetExecution(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		jsonutil.WriteErrorJSON(w, "Execution ID required")
		return
	}

	exec, err := h.repo.GetExecution(r.Context(), id)
	if err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, exec)
}

func (h *FlowHandler) ResumeExecution(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExecutionID string                 `json:"execution_id"`
		Overrides   map[string]interface{} `json:"overrides"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if err := h.runner.Resume(r.Context(), req.ExecutionID, req.Overrides); err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"status": "resumed"})
}

func (h *FlowHandler) CreateFlow(w http.ResponseWriter, r *http.Request) {
	var flow domain.Flow
	if err := json.NewDecoder(r.Body).Decode(&flow); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if err := h.repo.CreateFlow(r.Context(), &flow); err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusCreated, flow)
}

func (h *FlowHandler) GetFlow(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		// Try parsing from path if needed, but query is simpler for now
		jsonutil.WriteErrorJSON(w, "Flow ID required")
		return
	}

	flow, err := h.repo.GetFlow(r.Context(), id)
	if err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, flow)
}

func (h *FlowHandler) ListFlows(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	if zoneID == "" {
		jsonutil.WriteErrorJSON(w, "zone_id is required")
		return
	}

	flows, err := h.repo.ListFlows(r.Context(), zoneID)
	if err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, flows)
}

func (h *FlowHandler) UpdateFlow(w http.ResponseWriter, r *http.Request) {
	var flow domain.Flow
	if err := json.NewDecoder(r.Body).Decode(&flow); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if err := h.repo.UpdateFlow(r.Context(), &flow); err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, flow)
}

func (h *FlowHandler) BulkUpdateFlows(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs     []string `json:"ids"`
		Enabled bool     `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if err := h.repo.BulkUpdateFlowsEnabled(r.Context(), req.IDs, req.Enabled); err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

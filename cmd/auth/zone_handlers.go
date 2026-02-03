package main

import (
	"encoding/json"
	"net/http"

	"github.com/marwan562/fintech-ecosystem/internal/zone"
	"github.com/marwan562/fintech-ecosystem/internal/zone/domain"
	"github.com/marwan562/fintech-ecosystem/pkg/jsonutil"
)

type ZoneHandler struct {
	service *zone.Service
}

type CreateZoneRequest struct {
	OrgID string `json:"org_id"`
	Name  string `json:"name"`
	Mode  string `json:"mode"` // "test" or "live"
}

func (h *ZoneHandler) CreateZone(w http.ResponseWriter, r *http.Request) {
	var req CreateZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	mode := domain.Mode(req.Mode)
	if mode != domain.ModeTest && mode != domain.ModeLive {
		mode = domain.ModeTest
	}

	z, err := h.service.CreateZone(r.Context(), domain.CreateZoneParams{
		OrgID: req.OrgID,
		Name:  req.Name,
		Mode:  mode,
	})
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to create zone")
		return
	}

	jsonutil.WriteJSON(w, http.StatusCreated, z)
}

func (h *ZoneHandler) ListZones(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		jsonutil.WriteErrorJSON(w, "org_id query parameter required")
		return
	}

	zones, err := h.service.ListZones(r.Context(), orgID)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to list zones")
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, zones)
}

func (h *ZoneHandler) GetZone(w http.ResponseWriter, r *http.Request) {
	// In a real app we'd use a router with params like /zones/:id
	// For now we'll just check query param or path if we had chi/mux
	// Using standard library mux for now
	id := r.URL.Query().Get("id")
	if id == "" {
		jsonutil.WriteErrorJSON(w, "id query parameter required")
		return
	}

	z, err := h.service.GetZone(r.Context(), id)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to get zone")
		return
	}
	if z == nil {
		jsonutil.WriteErrorJSON(w, "Zone not found")
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, z)
}

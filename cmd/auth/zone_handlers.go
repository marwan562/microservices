package main

import (
	"encoding/json"
	"net/http"

	"github.com/sapliy/fintech-ecosystem/internal/zone"
	"github.com/sapliy/fintech-ecosystem/internal/zone/domain"
	"github.com/sapliy/fintech-ecosystem/pkg/jsonutil"
)

type ZoneHandler struct {
	service         *zone.Service
	templateService *zone.TemplateService
}

type CreateZoneRequest struct {
	OrgID        string `json:"org_id"`
	Name         string `json:"name"`
	Mode         string `json:"mode"` // "test" or "live"
	TemplateName string `json:"template_name"`
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
		OrgID:        req.OrgID,
		Name:         req.Name,
		Mode:         mode,
		TemplateName: req.TemplateName,
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

type BulkUpdateMetadataRequest struct {
	ZoneIDs  []string          `json:"zone_ids"`
	Metadata map[string]string `json:"metadata"`
}

func (h *ZoneHandler) BulkUpdateMetadata(w http.ResponseWriter, r *http.Request) {
	var req BulkUpdateMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	count, err := h.service.BulkUpdateMetadata(r.Context(), req.ZoneIDs, req.Metadata)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to update metadata")
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"updated_count": count,
	})
}

func (h *ZoneHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates := h.templateService.List()
	jsonutil.WriteJSON(w, http.StatusOK, templates)
}

func (h *ZoneHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateType := r.URL.Query().Get("type")
	if templateType == "" {
		jsonutil.WriteErrorJSON(w, "type query parameter required")
		return
	}

	template, err := h.templateService.Get(zone.TemplateType(templateType))
	if err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, template)
}

type ApplyTemplateRequest struct {
	ZoneID       string `json:"zone_id"`
	TemplateType string `json:"template_type"`
}

func (h *ZoneHandler) ApplyTemplate(w http.ResponseWriter, r *http.Request) {
	var req ApplyTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.ZoneID == "" {
		jsonutil.WriteErrorJSON(w, "zone_id is required")
		return
	}
	if req.TemplateType == "" {
		jsonutil.WriteErrorJSON(w, "template_type is required")
		return
	}

	result, err := h.templateService.Apply(r.Context(), req.ZoneID, zone.TemplateType(req.TemplateType))
	if err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, result)
}

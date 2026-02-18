package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sapliy/fintech-ecosystem/internal/zone"
	"github.com/sapliy/fintech-ecosystem/internal/zone/domain"
	"github.com/sapliy/fintech-ecosystem/pkg/jsonutil"
)

type ZoneHandler struct {
	service         *zone.Service
	templateService *zone.TemplateService
}

func NewZoneHandler(service *zone.Service, templateService *zone.TemplateService) *ZoneHandler {
	return &ZoneHandler{service: service, templateService: templateService}
}

func (h *ZoneHandler) CreateZone(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID        string `json:"org_id"`
		Name         string `json:"name"`
		Mode         string `json:"mode"`
		TemplateName string `json:"template_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	mode := domain.Mode(req.Mode)
	z, err := h.service.CreateZone(r.Context(), domain.CreateZoneParams{
		OrgID:        req.OrgID,
		Name:         req.Name,
		Mode:         mode,
		TemplateName: req.TemplateName,
	})
	if err != nil {
		jsonutil.WriteErrorJSON(w, fmt.Sprintf("Failed to create zone: %v", err))
		return
	}

	jsonutil.WriteJSON(w, http.StatusCreated, z)
}

func (h *ZoneHandler) GetZone(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	z, err := h.service.GetZone(r.Context(), id)
	if err != nil || z == nil {
		jsonutil.WriteErrorJSON(w, "Zone not found")
		return
	}
	jsonutil.WriteJSON(w, http.StatusOK, z)
}

func (h *ZoneHandler) BulkUpdateMetadata(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneIDs  []string               `json:"zone_ids"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}
	// Simplified logic for migration
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"status": "partial_success"})
}

func (h *ZoneHandler) ListZones(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	zones, err := h.service.ListZones(r.Context(), orgID)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to list zones")
		return
	}
	jsonutil.WriteJSON(w, http.StatusOK, zones)
}

func (h *ZoneHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateType := r.URL.Query().Get("type")
	tmpl, err := h.templateService.Get(zone.TemplateType(templateType))
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Template not found")
		return
	}
	jsonutil.WriteJSON(w, http.StatusOK, tmpl)
}

func (h *ZoneHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	tmpls := h.templateService.List()
	jsonutil.WriteJSON(w, http.StatusOK, tmpls)
}

func (h *ZoneHandler) ApplyTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneID       string `json:"zone_id"`
		TemplateName string `json:"template_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}
	if _, err := h.templateService.Apply(r.Context(), req.ZoneID, zone.TemplateType(req.TemplateName)); err != nil {
		jsonutil.WriteErrorJSON(w, err.Error())
		return
	}
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Template applied"})
}

package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SCIMUser represents a SCIM 2.0 user resource.
type SCIMUser struct {
	Schemas    []string    `json:"schemas"`
	ID         string      `json:"id,omitempty"`
	ExternalID string      `json:"externalId,omitempty"`
	UserName   string      `json:"userName"`
	Active     bool        `json:"active"`
	Name       SCIMName    `json:"name,omitempty"`
	Emails     []SCIMEmail `json:"emails,omitempty"`
	Meta       SCIMMeta    `json:"meta,omitempty"`
}

// SCIMName represents a user's name.
type SCIMName struct {
	Formatted  string `json:"formatted,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	GivenName  string `json:"givenName,omitempty"`
}

// SCIMEmail represents a user's email.
type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMMeta contains resource metadata.
type SCIMMeta struct {
	ResourceType string `json:"resourceType"`
	Created      string `json:"created,omitempty"`
	LastModified string `json:"lastModified,omitempty"`
	Location     string `json:"location,omitempty"`
	Version      string `json:"version,omitempty"`
}

// SCIMError represents a SCIM error response.
type SCIMError struct {
	Schemas  []string `json:"schemas"`
	Status   string   `json:"status"`
	Detail   string   `json:"detail,omitempty"`
	ScimType string   `json:"scimType,omitempty"`
}

// SCIMListResponse represents a list of SCIM resources.
type SCIMListResponse struct {
	Schemas      []string      `json:"schemas"`
	TotalResults int           `json:"totalResults"`
	StartIndex   int           `json:"startIndex"`
	ItemsPerPage int           `json:"itemsPerPage"`
	Resources    []interface{} `json:"Resources"`
}

// UserStore is the interface for managing users via SCIM.
type UserStore interface {
	CreateUser(ctx context.Context, user *SCIMUser) (*SCIMUser, error)
	GetUser(ctx context.Context, id string) (*SCIMUser, error)
	UpdateUser(ctx context.Context, id string, user *SCIMUser) (*SCIMUser, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, startIndex, count int) ([]*SCIMUser, int, error)
}

// SCIMHandler handles SCIM 2.0 requests.
type SCIMHandler struct {
	store UserStore
}

// NewSCIMHandler creates a new SCIM handler.
func NewSCIMHandler(store UserStore) *SCIMHandler {
	return &SCIMHandler{store: store}
}

// ServeHTTP implements http.Handler.
func (h *SCIMHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/scim+json")

	// Simple routing
	path := strings.TrimPrefix(r.URL.Path, "/scim/v2")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 {
		return // Discovery endpoint could go here
	}

	resource := parts[0]
	var id string
	if len(parts) > 1 {
		id = parts[1]
	}

	switch resource {
	case "Users":
		h.handleUsers(w, r, id)
	default:
		h.writeError(w, http.StatusNotFound, "Resource not found")
	}
}

func (h *SCIMHandler) handleUsers(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		switch r.Method {
		case http.MethodGet:
			h.listUsers(w, r)
		case http.MethodPost:
			h.createUser(w, r)
		default:
			h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getUser(w, r, id)
	case http.MethodPut:
		h.updateUser(w, r, id)
	case http.MethodDelete:
		h.deleteUser(w, r, id)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *SCIMHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination
	startIndex := 1
	count := 100
	// (Actual implementation would parse query params)

	users, total, err := h.store.ListUsers(r.Context(), startIndex, count)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := SCIMListResponse{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: total,
		StartIndex:   startIndex,
		ItemsPerPage: len(users),
		Resources:    make([]interface{}, len(users)),
	}

	for i, u := range users {
		resp.Resources[i] = u
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *SCIMHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var user SCIMUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	created, err := h.store.CreateUser(r.Context(), &user)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *SCIMHandler) getUser(w http.ResponseWriter, r *http.Request, id string) {
	user, err := h.store.GetUser(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "User not found")
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *SCIMHandler) updateUser(w http.ResponseWriter, r *http.Request, id string) {
	var user SCIMUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	updated, err := h.store.UpdateUser(r.Context(), id, &user)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *SCIMHandler) deleteUser(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.store.DeleteUser(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SCIMHandler) writeError(w http.ResponseWriter, status int, detail string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(SCIMError{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		Status:  fmt.Sprintf("%d", status),
		Detail:  detail,
	})
}

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"
)

const localNodeNameSettingKey = "local_node_name"

func (h *NodeHandler) resolveLocalNodeName() string {
	if h.db != nil {
		if name, err := h.db.GetPanelSetting(localNodeNameSettingKey); err == nil && strings.TrimSpace(name) != "" {
			return strings.TrimSpace(name)
		}
	}
	return "Local Master Node"
}

// decodeNodeName accepts exactly one name field in a bounded JSON object.
func decodeNodeName(w http.ResponseWriter, r *http.Request) (string, error) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4096))
	if err != nil || !utf8.Valid(body) {
		return "", fmt.Errorf("Provide a valid name object of at most 4096 bytes")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return "", fmt.Errorf("Provide a JSON object containing only name")
	}
	var name string
	seen := false
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil || key != "name" || seen {
			return "", fmt.Errorf("Provide exactly one name field")
		}
		if err := decoder.Decode(&name); err != nil {
			return "", fmt.Errorf("Name must be a string")
		}
		seen = true
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') {
		return "", fmt.Errorf("Provide a valid name object")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return "", fmt.Errorf("Provide exactly one name object")
	}
	// Reject controls even at the edges, before trimming ordinary whitespace.
	for _, char := range name {
		if unicode.IsControl(char) {
			return "", fmt.Errorf("Name must not contain control characters")
		}
	}
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return "", fmt.Errorf("Name must contain between 1 and 100 characters")
	}
	return name, nil
}

func (h *NodeHandler) RenameNode(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "Only administrators can rename nodes")
		return
	}
	name, err := decodeNodeName(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := r.PathValue("id")
	if id == "local" {
		err = h.db.SetPanelSetting(localNodeNameSettingKey, name)
	} else {
		err = h.db.RenameNode(id, name)
	}
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Node not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to rename node")
		return
	}
	logAudit(h.db, r, "node_rename", fmt.Sprintf("Renamed node %q to %q", id, name))
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "name": name})
}

package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/oathlesss/oathless-terminal/internal/commands"
)

// API handles HTTP API requests.
type API struct {
	registry *commands.Registry
}

// NewAPI creates a new API handler.
func NewAPI() *API {
	return &API{registry: commands.New()}
}

// CommandRequest is the JSON body for POST /api/command.
type CommandRequest struct {
	Command string `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// CommandResponse is the JSON response from command execution.
type CommandResponse struct {
	Output string `json:"output"`
	Type   string `json:"type"`
}

// HandleCommand processes a command execution request.
func (a *API) HandleCommand(w http.ResponseWriter, r *http.Request) {
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, CommandResponse{
			Output: "invalid request body",
			Type:   "error",
		})
		return
	}

	// Strip leading/trailing whitespace
	cmd := strings.TrimSpace(req.Command)
	if cmd == "" {
		writeJSON(w, http.StatusOK, CommandResponse{Output: "", Type: "text"})
		return
	}

	// If args weren't provided in the JSON, parse from the command string
	if len(req.Args) == 0 {
		parts := strings.Fields(req.Command)
		if len(parts) > 1 {
			cmd = parts[0]
			req.Args = parts[1:]
		}
	}

	result := a.registry.Execute(cmd, req.Args)
	log.Printf("command: %s %v → type=%s", cmd, req.Args, result.Type)

	writeJSON(w, http.StatusOK, CommandResponse{
		Output: result.Output,
		Type:   result.Type,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

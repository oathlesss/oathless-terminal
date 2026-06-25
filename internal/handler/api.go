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
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// CommandResponse is the JSON response from command execution.
type CommandResponse struct {
	Output string `json:"output"`
	Type   string `json:"type"`
}

// HandleCommands returns the list of available commands.
func (a *API) HandleCommands(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.registry.Available())
}

// HandleCommand processes a command execution request with pipe support.
func (a *API) HandleCommand(w http.ResponseWriter, r *http.Request) {
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, CommandResponse{
			Output: "invalid request body",
			Type:   "error",
		})
		return
	}

	raw := strings.TrimSpace(req.Command)
	if raw == "" {
		writeJSON(w, http.StatusOK, CommandResponse{Output: "", Type: "text"})
		return
	}

	// Check for pipe
	if strings.Contains(raw, "|") {
		result := a.executePipeline(raw)
		log.Printf("pipeline: %s → type=%s", raw, result.Type)
		writeJSON(w, http.StatusOK, CommandResponse{
			Output: result.Output,
			Type:   result.Type,
		})
		return
	}

	// Single command
	cmd, args := parseCommand(raw)
	result := a.registry.Execute(cmd, args)
	log.Printf("command: %s %v → type=%s", cmd, args, result.Type)

	writeJSON(w, http.StatusOK, CommandResponse{
		Output: result.Output,
		Type:   result.Type,
	})
}

func parseCommand(raw string) (string, []string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return "", nil
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	return parts[0], parts[1:]
}

func (a *API) executePipeline(raw string) commands.Response {
	stages := strings.Split(raw, "|")
	var input string

	for i, stage := range stages {
		cmd, args := parseCommand(strings.TrimSpace(stage))
		if cmd == "" {
			continue
		}

		var result commands.Response
		if i == 0 {
			// First stage: execute normally
			result = a.registry.Execute(cmd, args)
		} else {
			// Subsequent stages: pass previous output as input
			result = a.registry.ExecuteWithInput(cmd, args, input)
		}

		// Special types (clear, theme, matrix) propagate immediately
		if result.Type == "clear" || result.Type == "matrix" || strings.HasPrefix(result.Type, "theme:") {
			return result
		}

		// Error stops the pipeline
		if result.Type == "error" {
			return result
		}

		input = result.Output
	}

	return commands.Response{Output: input, Type: "text"}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

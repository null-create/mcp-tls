package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/null-create/mcp-tls/pkg/auth"
	"github.com/null-create/mcp-tls/pkg/mcp"
	"github.com/null-create/mcp-tls/pkg/util"
	"github.com/null-create/mcp-tls/pkg/validate"

	"github.com/google/uuid"
	"github.com/null-create/logger"
)

type Handlers struct {
	log          *logger.Logger
	usersManager auth.UsersManager
	toolManager  *mcp.ToolManager
}

func NewHandler() Handlers {
	return Handlers{
		log:          logger.NewLogger("API", uuid.NewString()),
		usersManager: auth.NewUsersManager(),
		toolManager:  mcp.NewToolManager("mcp-tls-tool-manager", "1.0.0", true),
	}
}

func (h *Handlers) errorMsg(w http.ResponseWriter, err error, statusCode int) {
	h.log.Error("%v", err)
	http.Error(w, err.Error(), statusCode)
}

func (h *Handlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	type HealthResponse struct {
		Status string `json:"status"`
	}

	err := json.NewEncoder(w).Encode(HealthResponse{
		Status: "ok",
	})
	if err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
	}
}

func (h *Handlers) LoadToolsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorMsg(w, errors.New("method not allowed"), http.StatusBadRequest)
		return
	}

	if err := h.toolManager.LoadTools(); err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
	}

	// send confirmation response
	json.NewEncoder(w).Encode(`{"message":"tools loaded"}`)
}

func (h *Handlers) ValidateToolHandler(w http.ResponseWriter, r *http.Request) {
	var tool mcp.Tool
	if err := json.NewDecoder(r.Body).Decode(&tool); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid tool JSON: "+err.Error())
		return
	}

	result := h.validate(&tool)

	util.WriteJSON(w, result)
}

func (h *Handlers) ValidateToolsHandler(w http.ResponseWriter, r *http.Request) {
	var tools []mcp.Tool
	if err := json.NewDecoder(r.Body).Decode(&tools); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid JSON array: "+err.Error())
		return
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = make([]mcp.ToolValidationResult, 0, len(tools))
	)

	for _, tool := range tools {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result := h.validate(&tool)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()

	}
	wg.Wait()

	util.WriteJSON(w, results)
}

func (h *Handlers) validate(tool *mcp.Tool) mcp.ToolValidationResult {
	origTool, err := h.toolManager.GetTool(tool.Name)
	if err != nil {
		h.log.Error("%v", err)
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: err.Error(),
		}
	}

	if tool.SecurityMetadata.Signature != origTool.SecurityMetadata.Signature ||
		tool.SecurityMetadata.Checksum != origTool.SecurityMetadata.Checksum {
		h.log.Error("signature or checksum mismatch")
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: "signature or checksum mismatch",
		}
	}

	// validate tool description
	err = validate.ValidateToolDescription(tool.Description)
	if err != nil {
		h.log.Error("tool description validation failed: %v", err)
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: err.Error(),
		}
	}

	// validate tool schema
	status, err := validate.ValidateToolInputSchema(tool, tool.Arguments)
	if err != nil {
		h.log.Error("tool input validation failed: %v", err)
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: err.Error(),
		}
	}
	if status == validate.StatusFailed {
		h.log.Error("%v", status)
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: "validation failed",
		}
	}

	h.log.Info("tool '%s' validated", tool.Name)
	return mcp.ToolValidationResult{
		Name:     tool.Name,
		Valid:    true,
		Checksum: tool.SecurityMetadata.Checksum,
	}
}

// Lists tools known to the server
func (h *Handlers) ListToolsHandler(w http.ResponseWriter, r *http.Request) {
	tools := h.toolManager.GetTools()
	if err := json.NewEncoder(w).Encode(tools); err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
	}
}

// Handles tool registration
func (h *Handlers) RegisterToolHandler(w http.ResponseWriter, r *http.Request) {
	var tool mcp.Tool
	if err := json.NewDecoder(r.Body).Decode(&tool); err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
		return
	}
	if tool.SecurityMetadata.IsEmpty() {
		h.errorMsg(w, errors.New("no security metadata found"), http.StatusBadRequest)
		return
	}

	// confirm checksums and signatures before registering
	err := validate.ValidateToolDescription(tool.Description)
	if err != nil {
		h.errorMsg(w, fmt.Errorf("tool registration failed. description invalid: %v", err), http.StatusBadRequest)
		return
	}

	cs, err := validate.GenerateToolChecksum(tool)
	if err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
		return
	}
	if cs != tool.SecurityMetadata.Checksum {
		csErr := ErrInvalidTool(fmt.Sprintf("checksum mismatch: given %v calculated %s", tool.SecurityMetadata.Checksum, cs))
		h.errorMsg(w, csErr, http.StatusBadRequest)
		return
	}

	fp, err := validate.GenerateSchemaFingerprint(tool.InputSchema)
	if err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
		return
	}
	if fp != tool.SecurityMetadata.Signature {
		fpErr := ErrInvalidTool(fmt.Sprintf("fingerprint mismatch: given %v calculated %s", tool.SecurityMetadata.Signature, fp))
		h.errorMsg(w, fpErr, http.StatusBadRequest)
		return
	}

	// tool security metadata has been validated. register tool.
	if err := h.toolManager.RegisterTool(tool); err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
		return
	}

	type Response struct {
		Msg string `json:"message"`
	}

	json.NewEncoder(w).Encode(Response{
		Msg: fmt.Sprintf("tool '%s' has been registered", tool.Name),
	})
}

// Gives a temporary token to the requestor to be able to register and valdiate tools
// Tokens last an hour by default
func (h *Handlers) TokenRequestHandler(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("userName")
	if userName == "" {
		h.errorMsg(w, errors.New("missing username"), http.StatusBadRequest)
		return
	}

	if !h.usersManager.HasUser(userName) {
		h.errorMsg(w, errors.New("register before requesting token"), http.StatusBadRequest)
		return
	}

	token, err := auth.CreateToken(userName, time.Hour)
	if err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
		return
	}

	type Token struct {
		Tok string `json:"token"`
	}

	err = json.NewEncoder(w).Encode(Token{Tok: token})
	if err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
	}
}

// Adds a new user to the session so they can be granted a token
func (h *Handlers) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("userName")
	if userName == "" {
		h.errorMsg(w, errors.New("missing username"), http.StatusBadRequest)
		return
	}

	// will be a no-op if the user is already registered
	h.usersManager.AddUser(userName)

	type RegisterResponse struct {
		Message string `json:"message"`
	}

	err := json.NewEncoder(w).Encode(RegisterResponse{
		Message: fmt.Sprintf("'%s' registered", userName),
	})
	if err != nil {
		h.errorMsg(w, err, http.StatusInternalServerError)
	}
}

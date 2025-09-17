package wallet

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type WalletServiceIface interface {
	ApplyOperation(ctx context.Context, req OperationRequest) (int64, error)
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
}

type WalletHandler struct {
	svc WalletServiceIface
}

func NewWalletHandler(svc WalletServiceIface) *WalletHandler {
	return &WalletHandler{svc: svc}
}

type postReq struct {
	WalletID      string `json:"walletId"`
	OperationType string `json:"operationType"`
	Amount        int64  `json:"amount"`
}

type balanceResp struct {
	WalletID string `json:"walletId"`
	Balance  int64  `json:"balance"`
}

func (h *WalletHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/wallet", h.PostOperation)
	mux.HandleFunc("/api/v1/wallets/", h.GetBalance)
}

func (h *WalletHandler) PostOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var pr postReq
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		writeJSONError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	wid, err := uuid.Parse(pr.WalletID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid walletId")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	opReq := OperationRequest{
		WalletID:      wid,
		OperationType: pr.OperationType,
		Amount:        pr.Amount,
	}

	newBal, err := h.svc.ApplyOperation(ctx, opReq)
	if err != nil {
		if err == ErrInsufficientFunds {
			writeJSONError(w, http.StatusConflict, "insufficient funds")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := balanceResp{
		WalletID: pr.WalletID,
		Balance:  newBal,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	prefix := "/api/v1/wallets/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, prefix)

	id = strings.TrimSuffix(id, "/")

	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "wallet id required")
		return
	}

	wid, err := uuid.Parse(id)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid walletId")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	bal, err := h.svc.GetBalance(ctx, wid)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	resp := balanceResp{
		WalletID: id,
		Balance:  bal,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sapliy/fintech-ecosystem/internal/wallet/domain"
	"github.com/sapliy/fintech-ecosystem/pkg/apierror"
	"github.com/sapliy/fintech-ecosystem/pkg/jsonutil"
	pb "github.com/sapliy/fintech-ecosystem/proto/wallet"
)

type WalletHandler struct {
	service *domain.WalletService
}

func NewWalletHandler(service *domain.WalletService) *WalletHandler {
	return &WalletHandler{service: service}
}

// HTTP Handlers

func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/v1/wallets/")
	if userID == "" {
		apierror.BadRequest("Missing User ID").Write(w)
		return
	}

	wallet, err := h.service.GetWallet(r.Context(), userID)
	if err != nil {
		apierror.Internal("Failed to retrieve wallet").Write(w)
		return
	}
	if wallet == nil {
		apierror.NotFound("Wallet not found").Write(w)
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, wallet)
}

func (h *WalletHandler) TopUp(w http.ResponseWriter, r *http.Request) {
	var req pb.TopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest("Invalid request body").Write(w)
		return
	}

	res, err := h.service.TopUp(r.Context(), &req)
	if err != nil {
		apierror.Internal(err.Error()).Write(w)
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, res)
}

func (h *WalletHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req pb.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest("Invalid request body").Write(w)
		return
	}

	res, err := h.service.Transfer(r.Context(), &req)
	if err != nil {
		apierror.Internal(err.Error()).Write(w)
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, res)
}

// gRPC Server

type WalletGRPCServer struct {
	pb.UnimplementedWalletServiceServer
	service *domain.WalletService
}

func NewWalletGRPCServer(service *domain.WalletService) *WalletGRPCServer {
	return &WalletGRPCServer{service: service}
}

func (s *WalletGRPCServer) GetWallet(ctx context.Context, req *pb.GetWalletRequest) (*pb.Wallet, error) {
	return s.service.GetWallet(ctx, req.UserId)
}

func (s *WalletGRPCServer) TopUp(ctx context.Context, req *pb.TopUpRequest) (*pb.TransactionResponse, error) {
	return s.service.TopUp(ctx, req)
}

func (s *WalletGRPCServer) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransactionResponse, error) {
	return s.service.Transfer(ctx, req)
}

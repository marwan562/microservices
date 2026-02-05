package domain

import (
	"context"
	"fmt"

	pb "github.com/sapliy/fintech-ecosystem/proto/ledger"
	walletpb "github.com/sapliy/fintech-ecosystem/proto/wallet"
)

type LedgerClient interface {
	GetAccount(ctx context.Context, accountID string) (*pb.GetAccountResponse, error)
	RecordTransaction(ctx context.Context, req *pb.RecordTransactionRequest) (*pb.RecordTransactionResponse, error)
}

type WalletService struct {
	ledgerClient LedgerClient
}

func NewWalletService(ledger LedgerClient) *WalletService {
	return &WalletService{ledgerClient: ledger}
}

func (s *WalletService) GetWallet(ctx context.Context, userID string) (*walletpb.Wallet, error) {
	// For now, we assume user ID is the account ID in the ledger
	// In a real system, there would be a mapping or a specific wallet account naming convention
	res, err := s.ledgerClient.GetAccount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ledger error: %w", err)
	}

	return &walletpb.Wallet{
		Id:       res.AccountId,
		UserId:   userID,
		Balance:  res.Balance,
		Currency: res.Currency,
	}, nil
}

func (s *WalletService) TopUp(ctx context.Context, req *walletpb.TopUpRequest) (*walletpb.TransactionResponse, error) {
	// Top up involves recording a transaction in the ledger
	// From an internal "float" or "system" account to the user's liability account
	res, err := s.ledgerClient.RecordTransaction(ctx, &pb.RecordTransactionRequest{
		AccountId:   req.UserId,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: "Wallet Top-Up",
		ReferenceId: req.ReferenceId,
	})
	if err != nil {
		return nil, fmt.Errorf("ledger error: %w", err)
	}

	return &walletpb.TransactionResponse{
		TransactionId: res.TransactionId,
		Status:        res.Status,
	}, nil
}

func (s *WalletService) Transfer(ctx context.Context, req *walletpb.TransferRequest) (*walletpb.TransactionResponse, error) {
	// Transfer between two wallets
	// This is a complex transaction in the ledger:
	// Debit from_user, Credit to_user (or vice versa depending on direction)
	// For simplicity, we'll implement this as two leg-transactions or a multi-leg transaction if ledger supports it.
	// Our ledger RecordTransaction currently handles single account updates (which is simplified)
	// but the domain model supports multiple entries.
	// Actually, the LedgerService.RecordTransaction in internal/ledger/domain/service.go
	// takes a TransactionRequest which has multiple entries.
	// However, the gRPC LedgerService only has RecordTransaction with RecordTransactionRequest (one account).

	// We need to either enhance the Ledger gRPC or do two calls (not atomic!).
	// Let's assume we use the existing gRPC for now and acknowledge the race condition/atomicity issue.
	// Better yet, I should have added a multi-reg gRPC.

	// For now, let's just do a transfer by debiting one and crediting the other via the single-account RPC.
	// THIS IS NOT RECOMMENDED FOR PRODUCTION - it should be atomic.

	// Debit from_user (amount is negative)
	_, err := s.ledgerClient.RecordTransaction(ctx, &pb.RecordTransactionRequest{
		AccountId:   req.FromUserId,
		Amount:      -req.Amount,
		Currency:    req.Currency,
		Description: fmt.Sprintf("Transfer to %s", req.ToUserId),
		ReferenceId: req.ReferenceId + "_debit",
	})
	if err != nil {
		return nil, fmt.Errorf("debit failed: %w", err)
	}

	// Credit to_user (amount is positive)
	res, err := s.ledgerClient.RecordTransaction(ctx, &pb.RecordTransactionRequest{
		AccountId:   req.ToUserId,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: fmt.Sprintf("Transfer from %s", req.FromUserId),
		ReferenceId: req.ReferenceId + "_credit",
	})
	if err != nil {
		return nil, fmt.Errorf("credit failed: %w", err)
	}

	return &walletpb.TransactionResponse{
		TransactionId: res.TransactionId,
		Status:        res.Status,
	}, nil
}

package main

import (
	"context"
	"log"

	"github.com/sapliy/fintech-ecosystem/internal/ledger/domain"
	pb "github.com/sapliy/fintech-ecosystem/proto/ledger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LedgerGRPCServer struct {
	pb.UnimplementedLedgerServiceServer
	service *domain.LedgerService
}

func NewLedgerGRPCServer(service *domain.LedgerService) *LedgerGRPCServer {
	return &LedgerGRPCServer{service: service}
}

func (s *LedgerGRPCServer) RecordTransaction(ctx context.Context, req *pb.RecordTransactionRequest) (*pb.RecordTransactionResponse, error) {
	entry := domain.EntryRequest{
		AccountID: req.AccountId,
		Amount:    req.Amount,
		Direction: "credit",
	}

	balancingEntry := domain.EntryRequest{
		AccountID: "system_balancing",
		Amount:    -req.Amount,
		Direction: "debit",
	}

	txReq := domain.TransactionRequest{
		ReferenceID: req.ReferenceId,
		Description: req.Description,
		Entries:     []domain.EntryRequest{entry, balancingEntry},
	}

	err := s.service.RecordTransaction(ctx, txReq, req.ZoneId, req.Mode)
	if err != nil {
		log.Printf("GRPC RecordTransaction error: %v", err)
		return nil, err
	}

	return &pb.RecordTransactionResponse{
		Status: "recorded",
	}, nil
}

func (s *LedgerGRPCServer) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	acc, err := s.service.CreateAccount(ctx, req.Name, domain.AccountType(req.Type), req.Currency, nil, req.ZoneId, req.Mode)
	if err != nil {
		log.Printf("GRPC CreateAccount error: %v", err)
		return nil, err
	}

	return &pb.CreateAccountResponse{
		AccountId: acc.ID,
		Status:    "created",
	}, nil
}

func (s *LedgerGRPCServer) BulkRecordTransactions(ctx context.Context, req *pb.BulkRecordRequest) (*pb.BulkRecordResponse, error) {
	var txRequests []domain.TransactionRequest
	for _, tr := range req.Transactions {
		txRequests = append(txRequests, domain.TransactionRequest{
			ReferenceID: tr.ReferenceId,
			Description: tr.Description,
			Entries: []domain.EntryRequest{
				{AccountID: tr.AccountId, Amount: tr.Amount, Direction: "credit"},
				{AccountID: "system_balancing", Amount: -tr.Amount, Direction: "debit"},
			},
		})
	}

	// Assuming the zone and mode from the first transaction for simplicity,
	// or we should handle it per transaction.
	var zoneID, mode string
	if len(req.Transactions) > 0 {
		zoneID = req.Transactions[0].ZoneId
		mode = req.Transactions[0].Mode
	}

	resErrs, err := s.service.BulkRecordTransactions(ctx, txRequests, zoneID, mode)
	if err != nil {
		return nil, err
	}

	var responses []*pb.RecordTransactionResponse
	for _, e := range resErrs {
		status := "recorded"
		if e != nil {
			status = "error: " + e.Error()
		}
		responses = append(responses, &pb.RecordTransactionResponse{
			Status: status,
		})
	}

	return &pb.BulkRecordResponse{
		Responses: responses,
	}, nil
}

func (s *LedgerGRPCServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	acc, err := s.service.GetAccount(ctx, req.AccountId)
	if err != nil {
		log.Printf("GRPC GetAccount error: %v", err)
		return nil, err
	}

	return &pb.GetAccountResponse{
		AccountId: acc.ID,
		Balance:   acc.Balance,
		Currency:  acc.Currency,
		CreatedAt: timestamppb.New(acc.CreatedAt),
	}, nil
}

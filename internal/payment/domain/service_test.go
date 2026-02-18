package domain

import (
	"context"
	"testing"
)

func TestPaymentService_ListPaymentIntents(t *testing.T) {
	ctx := context.Background()
	zoneID := "zone-123"

	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int
	}{
		{"positive limit", 10, 10},
		{"zero limit defaults to 50", 0, 50},
		{"negative limit defaults to 50", -1, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{
				ListPaymentIntentsFunc: func(ctx context.Context, zid string, limit int) ([]PaymentIntent, error) {
					if limit != tt.expectedLimit {
						t.Errorf("expected limit %d, got %d", tt.expectedLimit, limit)
					}
					return []PaymentIntent{}, nil
				},
				CreatePaymentIntentFunc: func(ctx context.Context, intent *PaymentIntent) error {
					return nil
				},
			}
			service := NewPaymentService(repo)
			_, err := service.ListPaymentIntents(ctx, zoneID, tt.inputLimit)
			if err != nil {
				t.Fatalf("ListPaymentIntents failed: %v", err)
			}
		})
	}
}

func TestPaymentService_CreatePaymentIntent(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{
		CreatePaymentIntentFunc: func(ctx context.Context, intent *PaymentIntent) error {
			return nil
		},
	}
	service := NewPaymentService(repo)
	err := service.CreatePaymentIntent(ctx, &PaymentIntent{Amount: 100, Currency: "USD", ZoneID: "z1"})
	if err != nil {
		t.Fatalf("CreatePaymentIntent failed: %v", err)
	}
}

func TestPaymentService_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{
		UpdateStatusFunc: func(ctx context.Context, id, status string) error {
			return nil
		},
	}
	service := NewPaymentService(repo)
	err := service.UpdateStatus(ctx, "id", "SUCCEEDED")
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}
}
func TestPaymentService_CreatePaymentIntent_Validation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		intent  *PaymentIntent
		wantErr bool
	}{
		{"Valid intent", &PaymentIntent{Amount: 100, Currency: "USD", ZoneID: "z1"}, false},
		{"Zero amount", &PaymentIntent{Amount: 0, Currency: "USD", ZoneID: "z1"}, true},
		{"Empty currency", &PaymentIntent{Amount: 100, Currency: "", ZoneID: "z1"}, true},
		{"Empty zone", &PaymentIntent{Amount: 100, Currency: "USD", ZoneID: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{
				CreatePaymentIntentFunc: func(ctx context.Context, intent *PaymentIntent) error {
					return nil
				},
			}
			service := NewPaymentService(repo)
			err := service.CreatePaymentIntent(ctx, tt.intent)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePaymentIntent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentService_UpdateStatus_Validation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"Valid status", "SUCCEEDED", false},
		{"Invalid status", "INVALID", true},
		{"Empty status", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{
				UpdateStatusFunc: func(ctx context.Context, id, status string) error {
					return nil
				},
			}
			service := NewPaymentService(repo)
			err := service.UpdateStatus(ctx, "id", tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

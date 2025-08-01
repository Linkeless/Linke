package command

import (
	"context"
	"fmt"
	"time"

	"linke/internal/payment/domain/aggregate"
	"linke/internal/payment/domain/repository"
	"linke/internal/payment/domain/valueobject"
	"linke/internal/shared/domain"
	sharedvo "linke/internal/shared/valueobject"
)

// PaymentCommandHandler handles payment commands
type PaymentCommandHandler struct {
	paymentRepo repository.PaymentRepository
	eventBus    domain.EventBus
}

// NewPaymentCommandHandler creates a new PaymentCommandHandler
func NewPaymentCommandHandler(
	paymentRepo repository.PaymentRepository,
	eventBus domain.EventBus,
) *PaymentCommandHandler {
	return &PaymentCommandHandler{
		paymentRepo: paymentRepo,
		eventBus:    eventBus,
	}
}

// CreatePayment handles the CreatePaymentCommand
func (h *PaymentCommandHandler) CreatePayment(ctx context.Context, cmd CreatePaymentCommand) (*CreatePaymentResult, error) {
	// Create value objects
	paymentNumber, err := valueobject.NewPaymentNumber(cmd.PaymentNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid payment number: %w", err)
	}
	
	// Create shared value objects directly
	invoiceID, err := sharedvo.NewInvoiceID(cmd.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}
	
	userID, err := sharedvo.NewUserIDFromUint(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	currency, err := valueobject.NewCurrency(cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}
	
	// Convert to shared currency for amount
	sharedCurrency, err := valueobject.ConvertToSharedCurrency(currency)
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %w", err)
	}
	
	amount, err := sharedvo.NewMoney(cmd.Amount, sharedCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	
	paymentMethod, err := valueobject.NewPaymentMethod(cmd.PaymentMethod)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method: %w", err)
	}
	
	paymentGateway, err := valueobject.NewPaymentGateway(cmd.PaymentGateway)
	if err != nil {
		return nil, fmt.Errorf("invalid payment gateway: %w", err)
	}
	
	// Check if payment number already exists
	exists, err := h.paymentRepo.ExistsByPaymentNumber(ctx, paymentNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check payment number existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("payment number already exists: %s", paymentNumber.String())
	}
	
	// Create payment aggregate
	payment, err := aggregate.NewPayment(
		paymentNumber,
		invoiceID,
		userID,
		amount,
		paymentMethod,
		paymentGateway,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}
	
	// Set optional fields
	if cmd.Notes != "" {
		payment.SetNotes(cmd.Notes)
	}
	if cmd.Metadata != "" {
		payment.SetMetadata(cmd.Metadata)
	}
	
	// Save payment
	if err := h.paymentRepo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}
	
	// Publish domain events
	if err := h.eventBus.Publish(ctx, payment.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}
	
	payment.ClearDomainEvents()
	
	// Convert shared amount back to payment domain amount for response
	sharedAmount := payment.Amount()
	paymentCurrency, err := valueobject.ConvertFromSharedCurrency(sharedAmount.Currency())
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency for result: %w", err)
	}
	
	resultAmount, err := valueobject.NewMoney(sharedAmount.Amount(), paymentCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to create result amount: %w", err)
	}
	
	return &CreatePaymentResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Amount:        resultAmount,
		Status:        payment.Status(),
		CreatedAt:     payment.CreatedAt(),
	}, nil
}

// UpdatePaymentDetails handles the UpdatePaymentDetailsCommand
func (h *PaymentCommandHandler) UpdatePaymentDetails(ctx context.Context, cmd UpdatePaymentDetailsCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	if err := payment.SetPaymentDetails(
		cmd.PaymentIntentID,
		cmd.PaymentURL,
		cmd.QRCodeURL,
		cmd.RedirectURL,
		cmd.ExpiresAt,
	); err != nil {
		return nil, fmt.Errorf("failed to set payment details: %w", err)
	}
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	// Publish domain events
	if err := h.eventBus.Publish(ctx, payment.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}
	
	payment.ClearDomainEvents()
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}

// SetGatewayTransaction handles the SetGatewayTransactionCommand
func (h *PaymentCommandHandler) SetGatewayTransaction(ctx context.Context, cmd SetGatewayTransactionCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	// Determine fee currency - start with payment currency (shared type)  
	sharedFeeCurrency := payment.Amount().Currency()
	if cmd.GatewayFeeCurrency != "" {
		// Convert to payment domain currency first, then to shared
		paymentFeeCurrency, err := valueobject.NewCurrency(cmd.GatewayFeeCurrency)
		if err != nil {
			return nil, fmt.Errorf("invalid gateway fee currency: %w", err)
		}
		sharedFeeCurrency, err = valueobject.ConvertToSharedCurrency(paymentFeeCurrency)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gateway fee currency: %w", err)
		}
	}
	
	// Create shared money for gateway fee
	sharedGatewayFee, err := sharedvo.NewMoney(cmd.GatewayFee, sharedFeeCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway fee: %w", err)
	}
	
	if err := payment.SetGatewayTransaction(cmd.TransactionID, sharedGatewayFee); err != nil {
		return nil, fmt.Errorf("failed to set gateway transaction: %w", err)
	}
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	// Publish domain events
	if err := h.eventBus.Publish(ctx, payment.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}
	
	payment.ClearDomainEvents()
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}

// ProcessPayment handles the ProcessPaymentCommand
func (h *PaymentCommandHandler) ProcessPayment(ctx context.Context, cmd ProcessPaymentCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	if err := payment.Process(); err != nil {
		return nil, fmt.Errorf("failed to process payment: %w", err)
	}
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	// Publish domain events
	if err := h.eventBus.Publish(ctx, payment.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}
	
	payment.ClearDomainEvents()
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}

// CompletePayment handles the CompletePaymentCommand
func (h *PaymentCommandHandler) CompletePayment(ctx context.Context, cmd CompletePaymentCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	if err := payment.Complete(cmd.WebhookData); err != nil {
		return nil, fmt.Errorf("failed to complete payment: %w", err)
	}
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	// Publish domain events
	if err := h.eventBus.Publish(ctx, payment.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}
	
	payment.ClearDomainEvents()
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}

// FailPayment handles the FailPaymentCommand
func (h *PaymentCommandHandler) FailPayment(ctx context.Context, cmd FailPaymentCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	if err := payment.Fail(cmd.Reason, cmd.WebhookData); err != nil {
		return nil, fmt.Errorf("failed to fail payment: %w", err)
	}
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	// Publish domain events
	if err := h.eventBus.Publish(ctx, payment.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}
	
	payment.ClearDomainEvents()
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}

// CancelPayment handles the CancelPaymentCommand
func (h *PaymentCommandHandler) CancelPayment(ctx context.Context, cmd CancelPaymentCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	if err := payment.Cancel(cmd.Reason); err != nil {
		return nil, fmt.Errorf("failed to cancel payment: %w", err)
	}
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	// Publish domain events
	if err := h.eventBus.Publish(ctx, payment.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}
	
	payment.ClearDomainEvents()
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}

// RefundPayment handles the RefundPaymentCommand
func (h *PaymentCommandHandler) RefundPayment(ctx context.Context, cmd RefundPaymentCommand) (*RefundPaymentResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	// Convert shared currency to payment domain currency for creating refund amount
	paymentCurrency, err := valueobject.ConvertFromSharedCurrency(payment.Amount().Currency())
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %w", err)
	}
	
	paymentRefundAmount, err := valueobject.NewMoney(cmd.RefundAmount, paymentCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid refund amount: %w", err)
	}
	
	// Convert to shared currency and create shared money for the aggregate method
	sharedCurrency, err := valueobject.ConvertToSharedCurrency(paymentCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %w", err)
	}
	
	sharedRefundAmount, err := sharedvo.NewMoney(cmd.RefundAmount, sharedCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared refund amount: %w", err)
	}
	
	refundNumber := valueobject.GenerateRefundNumber(payment.PaymentNumber())
	
	if err := payment.AddRefund(sharedRefundAmount, cmd.Reason, refundNumber); err != nil {
		return nil, fmt.Errorf("failed to add refund: %w", err)
	}
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	// Publish domain events
	if err := h.eventBus.Publish(ctx, payment.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}
	
	payment.ClearDomainEvents()
	
	// Convert shared total refunded amount back to payment domain
	sharedTotalRefunded := payment.RefundAmount()
	totalRefundedCurrency, err := valueobject.ConvertFromSharedCurrency(sharedTotalRefunded.Currency())
	if err != nil {
		return nil, fmt.Errorf("failed to convert total refunded currency: %w", err)
	}
	
	totalRefunded, err := valueobject.NewMoney(sharedTotalRefunded.Amount(), totalRefundedCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to create total refunded amount: %w", err)
	}
	
	return &RefundPaymentResult{
		PaymentID:       payment.ID(),
		PaymentNumber:   payment.PaymentNumber(),
		RefundAmount:    paymentRefundAmount, // Use the payment domain amount we created
		TotalRefunded:   totalRefunded,
		RefundReference: refundNumber,
		RefundedAt:      time.Now(),
	}, nil
}

// UpdatePaymentNotification handles the UpdatePaymentNotificationCommand
func (h *PaymentCommandHandler) UpdatePaymentNotification(ctx context.Context, cmd UpdatePaymentNotificationCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	payment.IncrementNotificationCount()
	if cmd.WebhookData != "" {
		payment.UpdateWebhookData(cmd.WebhookData)
	}
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}

// UpdatePaymentNotes handles the UpdatePaymentNotesCommand
func (h *PaymentCommandHandler) UpdatePaymentNotes(ctx context.Context, cmd UpdatePaymentNotesCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	payment.SetNotes(cmd.Notes)
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}

// UpdatePaymentMetadata handles the UpdatePaymentMetadataCommand
func (h *PaymentCommandHandler) UpdatePaymentMetadata(ctx context.Context, cmd UpdatePaymentMetadataCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	payment.SetMetadata(cmd.Metadata)
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}

// DeletePayment handles the DeletePaymentCommand
func (h *PaymentCommandHandler) DeletePayment(ctx context.Context, cmd DeletePaymentCommand) (*PaymentCommandResult, error) {
	paymentID, err := valueobject.NewPaymentID(cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}
	
	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	
	if err := payment.SoftDelete(); err != nil {
		return nil, fmt.Errorf("failed to delete payment: %w", err)
	}
	
	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	
	return &PaymentCommandResult{
		PaymentID:     payment.ID(),
		PaymentNumber: payment.PaymentNumber(),
		Status:        payment.Status(),
		UpdatedAt:     payment.UpdatedAt(),
	}, nil
}
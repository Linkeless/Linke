package valueobject

import (
	"testing"
)

func TestNewTicketStatus(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid open status",
			value:   "open",
			wantErr: false,
		},
		{
			name:    "valid in_progress status",
			value:   "in_progress",
			wantErr: false,
		},
		{
			name:    "valid resolved status",
			value:   "resolved",
			wantErr: false,
		},
		{
			name:    "valid closed status",
			value:   "closed",
			wantErr: false,
		},
		{
			name:    "invalid status",
			value:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty status",
			value:   "",
			wantErr: true,
		},
		{
			name:    "case insensitive",
			value:   "OPEN",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTicketStatus(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTicketStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTicketStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		to       string
		expected bool
	}{
		{
			name:     "open to in_progress",
			from:     "open",
			to:       "in_progress",
			expected: true,
		},
		{
			name:     "open to resolved",
			from:     "open",
			to:       "resolved",
			expected: true,
		},
		{
			name:     "resolved to closed",
			from:     "resolved",
			to:       "closed",
			expected: true,
		},
		{
			name:     "closed to resolved (invalid)",
			from:     "closed",
			to:       "resolved",
			expected: false,
		},
		{
			name:     "resolved to open (reopen)",
			from:     "resolved",
			to:       "open",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fromStatus := MustTicketStatus(tt.from)
			toStatus := MustTicketStatus(tt.to)
			
			result := fromStatus.CanTransitionTo(toStatus)
			if result != tt.expected {
				t.Errorf("CanTransitionTo() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNewTicketPriority(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		wantErr  bool
		expected int
	}{
		{
			name:     "low priority",
			value:    "low",
			wantErr:  false,
			expected: 1,
		},
		{
			name:     "normal priority",
			value:    "normal",
			wantErr:  false,
			expected: 2,
		},
		{
			name:     "critical priority",
			value:    "critical",
			wantErr:  false,
			expected: 5,
		},
		{
			name:    "invalid priority",
			value:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority, err := NewTicketPriority(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTicketPriority() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && priority.Level() != tt.expected {
				t.Errorf("NewTicketPriority().Level() = %v, expected %v", priority.Level(), tt.expected)
			}
		})
	}
}

func TestTicketPriority_RequiresImmediateAttention(t *testing.T) {
	tests := []struct {
		name     string
		priority string
		expected bool
	}{
		{
			name:     "low priority",
			priority: "low",
			expected: false,
		},
		{
			name:     "normal priority",
			priority: "normal",
			expected: false,
		},
		{
			name:     "urgent priority",
			priority: "urgent",
			expected: true,
		},
		{
			name:     "critical priority",
			priority: "critical",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := MustTicketPriority(tt.priority)
			result := priority.RequiresImmediateAttention()
			if result != tt.expected {
				t.Errorf("RequiresImmediateAttention() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNewTicketNumber(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid ticket number",
			value:   "TK-20240101-ABC123",
			wantErr: false,
		},
		{
			name:    "invalid format - missing prefix",
			value:   "20240101-ABC123",
			wantErr: true,
		},
		{
			name:    "invalid format - invalid date",
			value:   "TK-2024-ABC123",
			wantErr: true,
		},
		{
			name:    "invalid format - invalid suffix",
			value:   "TK-20240101-ABC",
			wantErr: true,
		},
		{
			name:    "empty value",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTicketNumber(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTicketNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
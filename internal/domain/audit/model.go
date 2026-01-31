package audit

import (
	"encoding/json"
	"time"

	"github.com/voidmaindev/go-template/internal/common/filter"
)

// Action constants for audit logging
const (
	ActionLoginSuccess    = "login_success"
	ActionLoginFailed     = "login_failed"
	ActionLogout          = "logout"
	ActionPasswordChange  = "password_change"
	ActionPasswordReset   = "password_reset"
	ActionRoleAssigned    = "role_assigned"
	ActionRoleRemoved     = "role_removed"
	ActionOAuthLinked     = "oauth_linked"
	ActionOAuthUnlinked   = "oauth_unlinked"
	ActionAccountLocked   = "account_locked"
	ActionUserCreated     = "user_created"
	ActionUserDeleted     = "user_deleted"
	ActionTokenRefresh    = "token_refresh"
	ActionEmailVerified   = "email_verified"
	ActionSelfRegistered  = "self_registered"
)

// AuditLog represents a security audit log entry
type AuditLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Timestamp  time.Time `gorm:"index;not null;default:CURRENT_TIMESTAMP" json:"timestamp"`
	UserID     *uint     `gorm:"index" json:"user_id,omitempty"`           // Nullable for pre-auth events
	Action     string    `gorm:"size:50;not null;index" json:"action"`     // login, logout, password_change, etc.
	Resource   string    `gorm:"size:50" json:"resource,omitempty"`        // user, role, identity
	ResourceID *uint     `json:"resource_id,omitempty"`                    // ID of affected resource
	IP         string    `gorm:"size:45" json:"ip,omitempty"`              // IPv4/IPv6
	UserAgent  string    `gorm:"size:500" json:"user_agent,omitempty"`     // Truncated if too long
	Success    bool      `gorm:"not null" json:"success"`
	Details    string    `gorm:"type:text" json:"details,omitempty"`       // JSON for extra context
}

// TableName returns the table name for GORM
func (AuditLog) TableName() string {
	return "audit_logs"
}

// FilterConfig enables filtering/sorting on list endpoints
func (AuditLog) FilterConfig() filter.Config {
	return filter.Config{
		TableName: "audit_logs",
		Fields: map[string]filter.FieldConfig{
			"id":         {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"timestamp":  {DBColumn: "timestamp", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"user_id":    {DBColumn: "user_id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"action":     {DBColumn: "action", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"resource":   {DBColumn: "resource", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"ip":         {DBColumn: "ip", Type: filter.TypeString, Operators: filter.StringOps, Sortable: false},
			"success":    {DBColumn: "success", Type: filter.TypeBool, Operators: filter.BoolOps, Sortable: true},
		},
	}
}

// AuditEntry is the input structure for creating audit logs
type AuditEntry struct {
	UserID     *uint
	Action     string
	Resource   string
	ResourceID *uint
	IP         string
	UserAgent  string
	Success    bool
	Details    map[string]any
}

// ToAuditLog converts an AuditEntry to an AuditLog model
func (e *AuditEntry) ToAuditLog() *AuditLog {
	log := &AuditLog{
		Timestamp:  time.Now(),
		UserID:     e.UserID,
		Action:     e.Action,
		Resource:   e.Resource,
		ResourceID: e.ResourceID,
		IP:         e.IP,
		UserAgent:  truncateString(e.UserAgent, 500),
		Success:    e.Success,
	}

	// Serialize details to JSON
	if e.Details != nil {
		if data, err := json.Marshal(e.Details); err == nil {
			log.Details = string(data)
		}
	}

	return log
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

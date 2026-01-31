package audit

import "time"

// AuditLogResponse is the API response for an audit log entry
type AuditLogResponse struct {
	ID         uint       `json:"id"`
	Timestamp  time.Time  `json:"timestamp"`
	UserID     *uint      `json:"user_id,omitempty"`
	Action     string     `json:"action"`
	Resource   string     `json:"resource,omitempty"`
	ResourceID *uint      `json:"resource_id,omitempty"`
	IP         string     `json:"ip,omitempty"`
	UserAgent  string     `json:"user_agent,omitempty"`
	Success    bool       `json:"success"`
	Details    string     `json:"details,omitempty"`
}

// ToResponse converts an AuditLog to an AuditLogResponse
func (a *AuditLog) ToResponse() *AuditLogResponse {
	return &AuditLogResponse{
		ID:         a.ID,
		Timestamp:  a.Timestamp,
		UserID:     a.UserID,
		Action:     a.Action,
		Resource:   a.Resource,
		ResourceID: a.ResourceID,
		IP:         a.IP,
		UserAgent:  a.UserAgent,
		Success:    a.Success,
		Details:    a.Details,
	}
}

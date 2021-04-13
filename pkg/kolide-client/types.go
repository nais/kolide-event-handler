package kolide_client

import "time"

type DeviceFailure struct {
	Id      int `json:"id"`
	CheckId int `json:"check_id"`
}

type DeviceOwner struct {
	Email string `json:"email"`
}

type Device struct {
	Id              int         `json:"id"`
	Name            string      `json:"name"`
	OwnedBy         string      `json:"owned_by"`
	Platform        string      `json:"platform"`
	LastSeenAt      time.Time   `json:"last_seen_at"`
	FailureCount    int         `json:"failure_count"`
	PrimaryUserName string      `json:"primary_user_name"`
	Serial          string      `json:"serial"`
	AssignedOwner   DeviceOwner `json:"assigned_owner"`
}

type Check struct {
	Tags []string `json:"tags"`
}

type Pagination struct {
	Next          string `json:"next"`
	NextCursor    string `json:"next_cursor"`
	CurrentCursor string `json:"current_cursor"`
	Count         int    `json:"count"`
}

type PaginatedResponse struct {
	Data       []interface{} `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

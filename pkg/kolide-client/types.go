package kolide_client

type DeviceFailure struct {
	Id      int `json:"id"`
	CheckId int `json:"check_id"`
}

type DeviceOwner struct {
	Email string `json:"email"`
}

type Device struct {
	Id     int         `json:"id"`
	Serial string      `json:"serial"`
	Owner  DeviceOwner `json:"assigned_owner"`
}

type Check struct {
	Tags []string `json:"tags"`
}

type ResponsePagination struct {
	Next          string `json:"next"`
	NextCursor    string `json:"next_cursor"`
	CurrentCursor string `json:"current_cursor"`
	Count         int    `json:"count"`
}

type DeviceFailures struct {
	Data       []DeviceFailure    `json:"data"`
	Pagination ResponsePagination `json:"pagination"`
}

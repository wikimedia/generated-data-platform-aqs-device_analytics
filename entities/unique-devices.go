package entities

// UniqueDevicesResponse represents a container for the unique devices resultset.
type UniqueDevicesResponse struct {
	Items []UniqueDevices `json:"items"`
}

// UniqueDevices represents one result from the unique devices resultset.
type UniqueDevices struct {
	Project       string `json:"project"`
	AccessSite    string `json:"access-site"`
	Granularity   string `json:"granularity"`
	Timestamp     string `json:"timestamp"`
	Devices       int    `json:"devices"`
	Offset        int    `json:"offset"`
	Underestimate int    `json:"underestimate"`
}

package entities

// UniqueDevicesResponse represents a container for the unique devices resultset.
type UniqueDevicesResponse struct {
	Items []UniqueDevices `json:"items"`
}

// UniqueDevices represents one result from the unique devices resultset.
type UniqueDevices struct {
	Project       string `json:"project" example:"en.wikipedia.org"` // Wikimedia project domain
	AccessSite    string `json:"access-site" example:"all-sites"`    // Method of access
	Granularity   string `json:"granularity" example:"daily"`        // Frequency of data
	Timestamp     string `json:"timestamp" example:"20220101"`       // Timestamp in YYYYMMDD format
	Devices       int    `json:"devices" example:"62614522"`         // Number of unique devices
	Offset        int    `json:"offset" example:"13127765"`
	Underestimate int    `json:"underestimate" example:"49486757"`
}

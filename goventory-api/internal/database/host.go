package database

// Inventory host. These can be edited if needed
type Host struct {
	ID        string  `json:"id"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
	HostGroup string `json:"host_group"`
}


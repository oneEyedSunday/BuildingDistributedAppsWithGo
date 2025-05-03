package registry

type Registration struct {
	ServiceName ServiceName
	ServiceURL  string
	// RequiredServices are services needed. Service A needs services [B, C]
	RequiredServices []ServiceName
	// ServiceUpdateURL is a URL the registry can communicate back to the requesting service on
	ServiceUpdateURL string
}

type ServiceName string

const (
	LogService     = ServiceName("LogService")
	GradingService = ServiceName("GradingService")
)

type patchEntry struct {
	Name ServiceName
	URL  string
}

type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}

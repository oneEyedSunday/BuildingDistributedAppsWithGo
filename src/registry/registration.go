package registry

type Registration struct {
	ServiceName ServiceName
	ServiceURL  string
	// RequiredServices are services a registration needs. For example while registering Service A, we could mark it as needing services [B, C]
	RequiredServices []ServiceName
	// ServiceUpdateURL is the URL the registry can communicate back to the requesting service on (to handle patchEntries for example).
	ServiceUpdateURL string
}

type ServiceName string

const (
	LogService     = ServiceName("LogService")
	GradingService = ServiceName("GradingService")
	TeacherPortal  = ServiceName("TeacherPortal")
)

type patchEntry struct {
	Name ServiceName
	URL  string
}

type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}

func (p patch) IsEmpty() bool {
	return len(p.Added) == 0 && len(p.Removed) == 0
}

package target

// TargetFile is a mapping of the config file to its parsed targets.
type TargetFile struct {
	config  *Config
	Targets map[string]*Target
}

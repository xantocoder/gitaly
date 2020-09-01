package praefect

// HealthChecker manages information of healthy nodes.
type HealthChecker interface {
	// HealthyNodes gets a list of healthy storages by their virtual storage.
	HealthyNodes() map[string][]string
}

// HealthCheckerFunc is an adapter to turn conforming functions in to a HealthChecker.
type HealthCheckerFunc func() map[string][]string

func (fn HealthCheckerFunc) HealthyNodes() map[string][]string { return fn() }

// StaticHealthChecker returns the nodes as always healthy.
type StaticHealthChecker map[string][]string

func (healthyNodes StaticHealthChecker) HealthyNodes() map[string][]string {
	return healthyNodes
}

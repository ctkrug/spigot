package spigot

// Loader is implemented by limiters that can report their current
// utilization without mutating state. It exists mainly so callers (and
// the browser simulator) can visualize load without depending on each
// algorithm's internal fields.
type Loader interface {
	// Load reports current utilization in [0, 1]: 0 means idle/empty
	// capacity, 1 means fully saturated.
	Load() float64
}

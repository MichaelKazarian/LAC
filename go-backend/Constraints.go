package main

import "fmt"

// Constraint defines a function type for global hardware safety checks.
// It returns nil if the state is safe, or an error if a fault is detected.
type Constraint func(c *Controller) error

// GetGlobalConstraints returns a list of all active safety locks.
func GetGlobalConstraints() []Constraint {
	return []Constraint{
		checkAirPressure,
		// сюди додаються нові аварії
	}
}

// checkAirPressure verifies the state of the pneumatic system pressure switch.
func checkAirPressure(c *Controller) error {
	c.state.mu.RLock()
	pressureSignal := c.state.Device10In[PinAirPressure]
	c.state.mu.RUnlock()

	// 0 means pressure dropped
	if pressureSignal == 0 {
		return fmt.Errorf("Низький тиск у пневмосистемі")
	}
	return nil
}

package main

import "fmt"

// Constraint — тип функції для глобальних перевірок апаратних аварій.
// Якщо все гаразд, повертає nil. Якщо є аварія — повертає error.
type Constraint func(c *Controller) error

// GetGlobalConstraints повертає список усіх активних захисних блокувань верстата
func GetGlobalConstraints() []Constraint {
	return []Constraint{
		checkAirPressure,
		// сюди додаються нові аварії
	}
}

// checkAirPressure перевіряє стан реле тиску пневмосистеми (Pin6)
func checkAirPressure(c *Controller) error {
	c.state.mu.RLock()
	pressureSignal := c.state.Device10In[PinAirPressure]
	c.state.mu.RUnlock()

	// Припустимо, 0 — тиск впав (нормально-замкнутий контакт розімкнувся через відсутність повітря)
	if pressureSignal == 0 {
		return fmt.Errorf("Низький тиск у пневмосистемі (Pin6)")
	}
	return nil
}

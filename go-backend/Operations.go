package main

import (
	"fmt"
	"time"
)

// RegisterOperations реєструє всі технологічні операції.
func RegisterOperations(r *OperationRegistry) {
	r.Add("operation1",  "Операція 1",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitStop2s)} })
	r.Add("operation2",  "Операція 2",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitStop2s)} })
	r.Add("operation3",  "Операція 3",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation4",  "Операція 4",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation5",  "Операція 5",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation6",  "Операція 6",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation7",  "Операція 7",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation8",  "Операція 8",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation9",  "Операція 9",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })

	r.Add("sync_mirror", "Дзеркалювання", func() []Step {
		return []Step{
			{
				Name: "Включення двигуна",
				Do:   func(c *Controller) { c.apply(func() { c.state.Device20Out[31] = 1 }) },
				Wait: waitMotorOn,
			},
			{
				Name: "Синхронізація",
				Do: func(c *Controller) {
					c.apply(func() {
						for i := 0; i < 31; i++ {
							c.state.Device20Out[i] = c.state.Device10In[i]
						}
					})
				},
				Wait: waitSyncMirror,
				Cleanup: func(c *Controller) {
					c.apply(func() {
						for i := 0; i < 31; i++ { c.state.Device20Out[i] = 0 }
					})
				},
			},
		}
	})

	r.Add("op_safety_stop", "Безпечна зупинка", func() []Step {
		return []Step{
			{
				Name: "Стоп",
				Do:   func(c *Controller) { c.apply(func() { c.state.Device20Out[3] = 1 }) },
				Wait: waitStop2s,
				Cleanup: func(c *Controller) {
					c.apply(func() {
						for i := 0; i < 32; i++ { c.state.Device20Out[i] = 0 }
					})
				},
			},
		}
	})

	r.Add("operation10", "Операція 10", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation11", "Операція 11", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation12", "Операція 12", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation13", "Операція 13", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation14", "Операція 14", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation15", "Операція 15", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation16", "Операція 16", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation17", "Операція 17", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation18", "Операція 18", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation19", "Операція 19", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation20", "Операція 20", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
}

// --- Wait-функції специфічні для операцій ---

// waitMotorOn — імітує очікування увімкнення мотору.
func waitMotorOn(c *Controller) StepResult {
	timeout := time.After(100 * time.Millisecond)
	tick := time.Tick(5 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return StepResult{Status: StepOK}
		case <-tick:
			c.state.mu.RLock()
			locked := c.state.IsSafetyLocked
			c.state.mu.RUnlock()
			if locked {
				return StepResult{Status: StepAbort, Message: "EmergencyStop"}
			}
		}
	}
}

// waitSyncMirror — протягом 5 секунд дзеркалює входи.
// Якщо сенсор потрапляє в аварійний діапазон — викликає EmergencyStop.
func waitSyncMirror(c *Controller) StepResult {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		c.state.mu.RLock()
		val := c.state.SensorValue / 2
		locked := c.state.IsSafetyLocked
		inputs := c.state.Device10In
		c.state.mu.RUnlock()

		if locked {
			return StepResult{Status: StepAbort, Message: "EmergencyStop already active"}
		}
		if val > 250 && val < 300 {
			EmergencyStop(c, fmt.Sprintf("Перевищено поріг сенсора: %d", val))
			return StepResult{Status: StepAbort, Message: "Sensor threshold exceeded"}
		}
		c.apply(func() {
			for i := 0; i < 31; i++ {
				c.state.Device20Out[i] = inputs[i]
			}
		})
		time.Sleep(2 * time.Millisecond)
	}
	return StepResult{Status: StepOK}
}

// --- Do-функції для заглушок ---

func stepItWorks(c *Controller) {
	fmt.Println("✅ Це працює")
}

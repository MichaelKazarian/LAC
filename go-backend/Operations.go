// Цей файл містить реєстрацію та реалізацію технологічних операцій.
//
// # Структура файлу
//
// Кожна нетривіальна операція організована у власний блок з чіткою структурою:
//
//	buildXxx()   — фабрика: повертає []Step, може містити локальні змінні-замикання.
//	stepXxx()    — конструктор одного Step: збирає Do/Wait/Cleanup в єдину структуру.
//	doXxx()      — виконує фізичну дію (миттєво, без очікувань).
//	waitXxx()    — очікує завершення фізичного процесу (може блокуватися).
//	cleanupXxx() — прибирає після кроку (аналог defer, не викликається при EmergencyStop).
//
// Щоб додати нову операцію:
//  1. Зареєструй її в RegisterOperations() через r.Add(id, displayName, buildXxx).
//  2. Реалізуй buildXxx() та відповідні step/do/wait/cleanup функції нижче.
//  3. Прості операції з одним кроком можна реєструвати inline через StepDoWait.
package main

import (
	"fmt"
	"time"
)

// RegisterOperations реєструє всі технологічні операції.
func RegisterOperations(r *OperationRegistry) {
	r.Add("operation1",  "Тест вих. 1",  buildTest1)
	r.Add("operation2",  "Операція 2",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitStop2s)} })
	r.Add("operation3",  "Операція 3",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation4",  "Операція 4",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation5",  "Операція 5",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation6",  "Операція 6",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation7",  "Операція 7",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation8",  "Операція 8",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation9",  "Операція 9",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("sync_mirror",    "Дзеркалювання",   buildSyncMirror)
	r.Add("op_safety_stop", "Безпечна зупинка", buildSafetyStop)
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

// =============================================================================
// sync_mirror
// =============================================================================

func buildTest1() []Step {
	return []Step{
		stepTestOutEnable(),
		stepTestOutDisable(),
	}
}


func buildSyncMirror() []Step {
	return []Step{
		stepMotorOn(),
		stepSyncMirror(),
	}
}

func stepTestOutEnable() Step {
	return Step{
		Name: "Тестове вмикання",
		Do:   doTestOutEnable,
    Wait: waitTime(4 * time.Second),
	}
}

func stepTestOutDisable() Step {
	return Step{
		Name: "Тестове вимикання",
		Do:   doTestOutDisable,
	}
}


func stepMotorOn() Step {
	return Step{
		Name: "Включення двигуна",
		Do:   doMotorOn,
		Wait: waitMotorOn,
	}
}

func stepSyncMirror() Step {
	return Step{
		Name:    "Синхронізація",
		Do:      doSyncMirror,
		Wait:    waitSyncMirror,
		Cleanup: cleanupSyncMirror,
	}
}

func doMotorOn(c *Controller) {
	c.apply(func() { c.state.Device20Out[OutMainMotor] = 1 })
}

func doSyncMirror(c *Controller) {
	c.apply(func() {
		for i := 0; i < OutMainMotor; i++ {
			c.state.Device20Out[i] = c.state.Device10In[i]
		}
	})
}

func doTestOutEnable(c *Controller) {
  c.apply(func() { c.state.Device20Out[OutTestPin11] = 1 })
}

func doTestOutDisable(c *Controller) {
  c.apply(func() { c.state.Device20Out[OutTestPin11] = 0 })
}

func cleanupSyncMirror(c *Controller) {
	c.apply(func() {
		for i := 0; i < OutMainMotor; i++ { c.state.Device20Out[i] = 0 }
	})
}

// =============================================================================
// op_safety_stop
// =============================================================================

func buildSafetyStop() []Step {
	return []Step{
		stepSafetyStop(),
	}
}

func stepSafetyStop() Step {
	return Step{
		Name:    "Стоп",
		Do:      doSafetyStop,
		Wait:    waitStop2s,
		Cleanup: cleanupSafetyStop,
	}
}

func doSafetyStop(c *Controller) {
	c.apply(func() { c.state.Device20Out[3] = 1 })
}

func cleanupSafetyStop(c *Controller) {
	c.apply(func() {
		for i := 0; i < 32; i++ { c.state.Device20Out[i] = 0 }
	})
}

// =============================================================================
// Wait-функції специфічні для операцій
// =============================================================================

// waitMotorOn — імітує очікування увімкнення мотору.
func waitMotorOn(c *Controller) StepResult {
  // 1. Тайм-аут: якщо мотор не розкрутився за 3 секунди — це помилка заліза
  timeout := time.After(3 * time.Second)
  ticker := time.NewTicker(20 * time.Millisecond)
  defer ticker.Stop()

  for {
    select {
    case <-timeout:
      return StepResult{Status: StepFail, Message: "Двигун не розкрутився (Timeout)"}
      
    case <-ticker.C:
      c.state.mu.RLock()
      // Читаємо реальний вхід з Device10In
      motorReady := c.state.Device10In[PinMotorReady] == 1
      locked := c.state.IsSafetyLocked
      c.state.mu.RUnlock()

      if locked {
        return StepResult{Status: StepAbort, Message: "Зупинено через EmergencyStop"}
      }

      if motorReady {
        return StepResult{Status: StepOK} // Мотор готовий, йдемо до наступного кроку
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
		val := c.state.EncoderValue / 2
		locked := c.state.IsSafetyLocked
		inputs := c.state.Device10In
		c.state.mu.RUnlock()

		if locked {
			return StepResult{Status: StepAbort, Message: "EmergencyStop already active"}
		}
		if val > 250 && val < 300 {
			c.Stop(fmt.Sprintf("Перевищено поріг сенсора: %d", val))
			return StepResult{Status: StepAbort, Message: "Sensor threshold exceeded"}
		}
		c.apply(func() {
			for i := 0; i < OutMainMotor; i++ {
				c.state.Device20Out[i] = inputs[i]
			}
		})
		time.Sleep(2 * time.Millisecond)
	}
	return StepResult{Status: StepOK}
}

// waitTime повертає функцію очікування для заданої тривалості.
// Використовується як: Wait: waitTime(2 * time.Second)
func waitTime(duration time.Duration) func(c *Controller) StepResult {
	return func(c *Controller) StepResult {
		deadline := time.Now().Add(duration)

		for time.Now().Before(deadline) {
			// 1. Перевірка Safety Lock (Emergency Stop)
			c.state.mu.RLock()
			locked := c.state.IsSafetyLocked
			c.state.mu.RUnlock()

			if locked {
				return StepResult{
					Status:  StepAbort,
					Message: "Очікування перервано: Safety Lock",
				}
			}

			// 2. Коротка пауза, щоб не перевантажувати CPU
			// 20-50мс достатньо для оперативності в промислових задачах
			time.Sleep(25 * time.Millisecond)
		}

		return StepResult{Status: StepOK}
	}
}

// =============================================================================
// Do-функції для заглушок
// =============================================================================

func stepItWorks(c *Controller) {
	fmt.Println("✅ Це працює")
}

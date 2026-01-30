package main

import (
	"fmt"
	"time"
)

type Controller struct {
	hw    HardwareService
	state *HardwareState
}

func NewController(hw HardwareService, state *HardwareState) *Controller {
	return &Controller{hw: hw, state: state}
}

// Run — основний цикл контролера
func (c *Controller) Run() {
	fmt.Println("🚀 Контролер логіки запущено")
	var lastInputs [32]uint16
	firstRun := true

	for {
		start := time.Now()

		// Читаємо дані через абстрактний сервіс
		sensor, inputs, err := c.hw.Read()

		if err == nil {
			c.updateSlave3(sensor, true)
			c.updateSlave10(&inputs, true)

			// Виводимо результати в консоль (опціонально)
			// c.showResults(inputs) 

			// Логіка синхронізації зі Slave 20
			if inputs != lastInputs || firstRun {
				if errWrite := c.hw.Write(inputs); errWrite == nil {
					lastInputs = inputs
					firstRun = false
				}
			}
		} else {
			c.updateSlave10(nil, false)
			c.updateSlave3(0, false)
			fmt.Printf("[%s] Помилка читання: %v\n", time.Now().Format("15:04:05"), err)
		}

		c.updateCycleTime(time.Since(start).Milliseconds())
		// time.Sleep(5 * time.Millisecond)
	}
}

func (c *Controller) updateSlave3(val uint16, online bool) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.SensorValue = val
	c.state.IsOnline3 = online
	c.state.LastUpdate = time.Now()
}

func (c *Controller) updateSlave10(data *[32]uint16, online bool) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.IsOnline10 = online
	if online && data != nil {
		c.state.Device10In = *data
	}
	c.state.LastUpdate = time.Now()
}

func (c *Controller) updateCycleTime(ms int64) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.ReadCycleMs = ms
}

func (c *Controller) showResults(data [32]uint16) {
	// r0, r1 можна отримати назад через Pack32, якщо вони потрібні для друку
	r0, r1 := Pack32(&data)
	fmt.Printf("Reg0: %016b | Reg1: %016b\n", r0, r1)
	for i := 0; i < 32; i++ {
		fmt.Printf("%d:%d ", i, data[i])
		if (i+1)%16 == 0 {
			fmt.Println()
		}
	}
}

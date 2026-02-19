package main

import (
    "fmt"
    "time"
)

type MockPowerControl struct {
    enabled bool
}

func (m *MockPowerControl) EnableOutputsPower() error {
    if m.enabled {
        return nil
    }
    m.enabled = true
    fmt.Printf("[PWR] [%s] DO board POWER ON\n", time.Now().Format("15:04:05"))
    return nil
}

func (m *MockPowerControl) DisableOutputsPower() error {
    if !m.enabled {
        return nil
    }
    m.enabled = false
    fmt.Printf("[PWR] [%s] DO board POWER OFF (FAIL-SAFE)\n", time.Now().Format("15:04:05"))
    return nil
}

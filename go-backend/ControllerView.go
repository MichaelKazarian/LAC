package main

import "time"

// SystemView — це ЄДИНЕ, що можна показувати назовні.
// Тому не має mutex, логіки і не використовується для керування.
type SystemView struct {
	Mode            ControlMode `json:"mode"`
	IsPaused        bool        `json:"is_paused"`
	IsSafetyLocked  bool        `json:"is_safety_locked"`
	StopReason      string      `json:"stop_reason"`
	ActiveOperation string      `json:"active_operation"`
	Counter         int         `json:"counter"`

	EncoderValue    uint16      `json:"encoder_value"`
	Device10In      [32]uint16  `json:"device10_in"`
  Device20Out     [32]uint16  `json:"device20_out"`

	IsEncoderOnline bool        `json:"is_encoder_online"`
	IsInputsOnline  bool        `json:"is_inputs_online"`
	IsOutputsOnline bool        `json:"is_outputs_online"`

	ReadCycleMs     int64       `json:"read_cycle_ms"`
	LastUpdate      time.Time   `json:"last_update"`

  OpsList         [][]string  `json:"-"`
}

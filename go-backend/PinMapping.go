// Package main містить константи для відображення (mapping) фізичних входів та виходів
// Modbus-периферії на внутрішні структури даних контролера.
//
// Файл PinMapping.go є єдиним джерелом істини для адресації I/O.
// Будь-які зміни в електричній схемі підключення мають відображатися тут.
//
// Адресація розділена за типами Slave-пристроїв:
//   - Slave 10 (Inputs): Дискретні входи (датчики, кнопки, зворотний зв'язок)
//   - Slave 20 (Outputs): Дискретні виходи (реле, пускачі, сигналізація)
package main

import (
	"fmt"
)

// --- Inputs Mapping (Slave 10 / Device10In) ---
const (
	// PinMotorReady — сигнал готовності частотного перетворювача (Drive Ready).
	PinMotorReady = 5
)

// --- Outputs Mapping (Slave 20 / Device20Out) ---
const (
	// OutMainMotor — керуючий сигнал на котушку пускача головного двигуна.
	// Вимикається примусово через IsSafetyLocked.
	OutMainMotor = 31
  OutTestPin      = 16
)

// --- Modbus Addresses ---
const (
	AddrEncoder = 3  // Енкодер
	AddrInputs  = 10 // Дискретні входи
	AddrOutputs = 20 // Релейні виходи
)

// PinNames дозволяє отримати текстову назву піна за його індексом.
// Використовується для логування подій та відображення стану в UI.
var PinNames = map[int]string{
	// Входи
	PinMotorReady:   "Готовність двигуна",

	// Виходи
	OutMainMotor: "Головний двигун",
  OutTestPin:   "Тестовий пін",
}

// GetPinName повертає назву піна або "Unknown", якщо пін не описаний.
func GetPinName(pin int) string {
	if name, ok := PinNames[pin]; ok {
		return name
	}
	return fmt.Sprintf("Pin %d", pin)
}

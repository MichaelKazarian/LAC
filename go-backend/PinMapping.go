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
	PinMotorReady = 0
)

// --- Outputs Mapping (Slave 20 / Device20Out) ---
const (
	// OutMainMotor — керуючий сигнал на котушку пускача головного двигуна.
	// Вимикається примусово через IsSafetyLocked.
	OutMainMotor      = 31
  OutDrivePower     = 10
  OutSpindleMotor   = 11
)

// --- Modbus Addresses ---
const (
	AddrEncoder = 3  // Енкодер
	AddrInputs  = 10 // Дискретні входи
	AddrOutputs = 20 // Релейні виходи
)

// PinNamesIn описує назви вхідних сигналів (Device 10)
var PinNamesIn = map[int]string{
    PinMotorReady: "Готовність двигуна",
    // Додайте інші входи тут
}

// PinNamesOut описує назви вихідних сигналів (Device 20)
var PinNamesOut = map[int]string{
    OutMainMotor:  "Головний двигун",
    OutDrivePower:  "Пускач живлення приводів",
    OutSpindleMotor:  "Включення мотора шпінделя",
}

// GetPinName повертає назву піна або "Unknown", якщо пін не описаний.
func GetPinName(pin int, isOutput bool) string {
    mapping := PinNamesIn
    prefix := "Вхід"
    if isOutput {
        mapping = PinNamesOut
        prefix = "Вихід"
    }

    if name, ok := mapping[pin]; ok {
        return name
    }

    return fmt.Sprintf("%s [%d]", prefix, pin)
}

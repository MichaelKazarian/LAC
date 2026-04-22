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
  Pin5          = 5
  Pin17         = 17
  Pin18         = 18 
)

// --- Outputs Mapping (Slave 20 / Device20Out) ---
const (
	// OutMainMotor — керуючий сигнал на котушку пускача головного двигуна.
	// Вимикається примусово через IsSafetyLocked.
  OutDrivePower     = 10
  OutSpindleMotor   = 11
  OutTestPin12      = 12
  OutTestPin13      = 13
  OutTestPin14      = 14
  OutTestPin15      = 15
  OutTestPin16      = 16
  OutTestPin17      = 17
  OutTestPin18      = 18
  OutTestPin19      = 19
  OutTestPin20      = 20
  OutTestPin21      = 21
  OutTestPin22      = 22
  OutTestPin23      = 23
  OutTestPin24      = 24
  OutTestPin25      = 25
  OutTestPin26      = 26
  OutTestPin27      = 27
  OutTestPin28      = 28
  OutTestPin29      = 29
  OutTestPin30      = 30
  OutTestPin31      = 31
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
  Pin5: "Реле тиску пневмоситеми піля пристрою плавного пуску",
  Pin17: "Відвод інструмента вихідне - на осі шпінделя",
  Pin18: "Відвод інструмента - відведено",
    // Додайте інші входи тут
}

// PinNamesOut описує назви вихідних сигналів (Device 20)
var PinNamesOut = map[int]string {
    OutDrivePower:  "Пускач живлення приводів", // depends OutDrivePower
    OutSpindleMotor:  "Включення мотора шпінделя", // depends OutDrivePower
    OutTestPin12:     "Вкл. мотора поливної змазки", //lubricant, depends OutDrivePower
    OutTestPin13:     "Індикація старту", // not depends2
    OutTestPin14:     "Індикація стопу",  // not depends2
    OutTestPin15:     "Зелений факел",  // not depends2
    OutTestPin16:     "Червоний факел",  // not depends2
    OutTestPin17:     "вкл.частотним перетвоювачем", //depends OutDrivePower
    OutTestPin18:     "ПЧВ швидкість 1", //depends OutDrivePower
    OutTestPin19:     "ПЧВ швидкість 2", //depends OutDrivePower
    OutTestPin20:     "ПЧВ реверс", //depends OutDrivePower and pin18/19
    OutTestPin21:     "Вловлювач", // ловить деталь,not depends
    OutTestPin22:     "Виштовхувач", // виштовхує деталь, not depends
    OutTestPin23:     "Відвод інструменту", // not depends, switch In 17-18
    OutTestPin24:     "Плавний пуск пневматики", // not depends, In 5 On
    OutTestPin25:     "Патрон шпінделя відкр.", // not depends
    OutTestPin26:     "Завантажувач до осі шпінделя", // Залежить від положення Out23
    OutTestPin27:     "Заштовхувач заготовки вперед", // Залежить від положення Out23, Out26
    OutTestPin28:     "Лоток відсікача відчинено", // Завантажує деталі в лоток заштовхувача (Out26) Датчик не працює
    OutTestPin29:     "Вхідний магазин відсікача відкритий", // Клапан відпацював, проблема з пневматикою.
    OutTestPin30:     "Продув шпінделя",
    OutTestPin31:     "Дозування змазки", //Вмикати періодично
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

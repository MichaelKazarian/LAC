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
	// PinMotorReady — готовністґьі частотного перетворювача (Drive Ready)
	// PinMotorReady = 12
  Pin3          = 3
  Pin4          = 4
  Pin5          = 5
  Pin6          = 6
  Pin7          = 7
  Pin8          = 8
  Pin9          = 9
  Pin10         = 10
  Pin11         = 11
  Pin12         = 12
  Pin13         = 13
  Pin14         = 14
  PinUnloaderHome         = 15
  PinUnloaderAxis         = 16
  PinToolAxis         = 17
  PinToolHome         = 18
  Pin19         = 19
  PinLoaderHome         = 20
  PinLoaderAxis         = 21
  PinPusherAxis         = 22
  PinPusherHome         = 23
  PinPartInLoader         = 24
  Pin25         = 25
  Pin26         = 26
  PinTrayGateHome         = 27
  PinTrayGateOpen         = 28
  PinMagShutterHome         = 29
  Pin30         = 30
  Pin31         = 31
)

// --- Outputs Mapping (Slave 20 / Device20Out) ---
const (
	// OutMainMotor — сигнал на пускач головного двигуна
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
  OutUnloader      = 21
  OutEjector      = 22
  OutTool      = 23
  OutTestPin24      = 24
  OutCollet      = 25
  OutLoader      = 26
  OutPusher      = 27
  OutTrayGateOpen      = 28
  OutMagShutterOpen      = 29
  OutAirBlast      = 30
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
  //PinMotorReady: "Готовність двигуна",
  Pin3: "Датчик дозованої змазки", // TODO: фізично не зайдено, якщо не вмикається видавати warning
  Pin4: "Реле тиску поливної змазки", // TODO: фізично не зайдено, якщо не вмикається видавати warnin
  Pin5: "Реле тиску пневмоситеми після пристрою плавного пуску",
  Pin6: "Реле тиску пневмосистеми",
  Pin7: "Вкл. вперед",
  Pin8: "Викл. назад",
  Pin9: "Аварійна зупинка", // Нормально замкнутий
  Pin10: "Блок-контакт автомата мотора шпінделя",
  Pin11: "Блок-контакт автомата привода подачі",
  Pin12: "Готовність привода подачі",
  Pin13: "Блок-контакт автомата вентилятора",
  Pin14: "Блок-контакт автомата мотора змазки",
  PinUnloaderHome: "Вигружач вихідне",
  PinUnloaderAxis: "Вигружач на осі шп.",
  PinToolAxis: "Інструмент: на осі",
  PinToolHome: "Інструмент: вихідне",
  Pin19: "Патрон шп. зажато",    // фізично не зайдено
  PinLoaderHome: "Завантажувач вих.",
  PinLoaderAxis: "Завантажувач на осі шп.",
  PinPusherAxis: "Заштовхувач: на осі (Axis)",
  PinPusherHome: "Заштовхувач: вихідне (Home)", // вих. положення в цангі
  PinPartInLoader: "Датчик заготовки в завантажувачі",
  Pin25: "Датчик заготовки на вих магазина",
  Pin26: "Патрон шп. розажато",  // фізично не зайдено
  PinTrayGateHome: "Вх. лоток відсікача вих",
  PinTrayGateOpen: "Вх. лоток відсікача відкр.",
  PinMagShutterHome: "Магазин відсікач. вих.",
  Pin30: "Вмикач живлення приводів",
  Pin31: "Вимикач живлення приводів",
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
    OutUnloader:     "Вивантажувач", // ловить деталь,not depends
    OutEjector:     "Виштовхувач (Ejector)", // виштовхує деталь, not depends
    OutTool:     "Інструмент", // not depends, switch In 17-18
    OutTestPin24:     "Плавний пуск пневматики", // not depends, In 5 On
    OutCollet:     "Цанга", // not depends
    OutLoader:     "Завантажувач до осі шпінделя", // Залежить від положення Out23
    OutPusher:     "Заштовхувач заготовки", // Залежить від положення Out23, Out26
    OutTrayGateOpen:     "Лоток відсікача відчинено", // Завантажує деталі в лоток заштовхувача (Out26) Датчик не працює
    OutMagShutterOpen:     "Вхідний магазин відсікача відкритий",
    OutAirBlast:     "Продув шпінделя",
    OutTestPin31:     "Дозування змазки", //Вмикати періодично TODO it temporary replace OutMainMotor fix it
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

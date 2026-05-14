package main

var (
  catalog     = make(map[string]OperationInfo)
  orderedKeys []string
)

func InitCatalog() {
  catalog = make(map[string]OperationInfo)
  orderedKeys = []string{}

  add := func(id, name string, build func() []Step) {
    catalog[id] = OperationInfo{ID: id, DisplayName: name, Build: build}
    orderedKeys = append(orderedKeys, id)
  }
  add("op_drive_power_on",  "Живлення приводів УВІМК", buildDrivePowerOn)
  add("op_drive_power_off", "Живлення приводів ВИМК", buildDrivePowerOff)
  add("op_mag_shutter",     "Завантаження магазину", buildMagShutter)
  add("op_tray_move",       "Крок лотка",            buildTrayMove)
  add("op_tray_move_auto",  "Переміщення лотка",     buildTrayAutoFill)
  add("op_loader",          "Цикл",                  buildLoader)
  add("op_spindle_on",      "Старт шпінделя",        buildSpindleOn)
  add("op_spindle_off",     "Стоп шпінделя",         buildSpindleOff)
  add("sync_mirror",         "Дзеркалювання",   buildSyncMirror)
  add("op_safety_stop",      "Безпечна зупинка", buildSafetyStop)
  add("op_move_to_safe_pos", "Розжати",          buildMoveToSafePosition)
  add("op_vfd_speed_1",     "ПЧВ: Швидкість 1",       buildVFDSpeed1)
  add("op_vfd_speed_2",     "ПЧВ: Швидкість 2",       buildVFDSpeed2)
  add("op_vfd_reverse",     "ПЧВ: Реверс",   buildVFDReverse)
  add("op_vfd_stop",        "ПЧВ: СТОП",   buildVFDStop)
}

func GetAutoModeConfig() AutoModeConfig {
  c := catalog
  return AutoModeConfig{
    Before: []OperationInfo{
      c["op_drive_power_on"],
      c["op_spindle_on"],
    },
    Main:   []OperationInfo{
      c["op_tray_move_auto"],
      c["op_loader"],
    },
    After:  []OperationInfo{
      c["op_spindle_off"],
    },
  }
}

func GetManualConfig() []OperationInfo {
  res := make([]OperationInfo, 0, len(orderedKeys))
  for _, id := range orderedKeys {
    res = append(res, catalog[id])
  }
  return res
}

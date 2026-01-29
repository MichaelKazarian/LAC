package main

// HardwareService - інтерфейс для роботи з обладнанням
type HardwareService interface {
  // GetState() *HardwareState
  Run()
  Close()
}

// Pack32 перетворює масив із 32 значень у два регістри uint16
func Pack32(data *[32]uint16) (uint16, uint16) {
    var reg0, reg1 uint16
    for i := 0; i < 16; i++ {
        if data[i] > 0 { reg0 |= (1 << i) }
        if data[i+16] > 0 { reg1 |= (1 << i) }
    }
    return reg0, reg1
}

// Unpack32 розпаковує два регістри uint16 назад у масив із 32 елементів
func Unpack32(reg0, reg1 uint16) [32]uint16 {
    var data [32]uint16
    for i := 0; i < 16; i++ {
        data[i] = (reg0 >> i) & 1
        data[i+16] = (reg1 >> i) & 1
    }
    return data
}

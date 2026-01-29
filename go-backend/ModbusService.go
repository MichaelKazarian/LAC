package main

import (
    "encoding/binary"
    "fmt"
    "log"
    "time"
    "github.com/goburrow/modbus"
)

type ModbusService struct {
    client  modbus.Client
    handler *modbus.RTUClientHandler
    state   *HardwareState
}

func NewModbusService(device string, baud int, state *HardwareState) *ModbusService {
    handler := modbus.NewRTUClientHandler(device)
    handler.BaudRate = baud
    handler.DataBits = 8
    handler.Parity = "N"
    handler.StopBits = 1
    handler.Timeout = 200 * time.Millisecond

    err := handler.Connect()
    if err != nil {
        log.Fatalf("❌ Помилка ініціалізації порту: %v", err)
    }
    fmt.Println("✅ Modbus порт успішно відкрито")
    return &ModbusService{client: modbus.NewClient(handler), handler: handler, state: state}
}

func (ms *ModbusService) Run() {
    firstRun := true
    var lastReg0, lastReg1 uint16
    fmt.Println("🚀 Цикл опитування Modbus запущено")

    for {
        start := time.Now()

        // --- Опитування Slave 3 ---
        ms.handler.SlaveId = 3
        res3, err3 := ms.client.ReadHoldingRegisters(0, 1)
        if err3 == nil {
            val := uint16(res3[0])<<8 | uint16(res3[1])
            ms.state.UpdateSlave3(val, true)
        } else {
            ms.state.UpdateSlave3(0, false)
        }
        time.Sleep(2 * time.Millisecond)

        // --- Опитування Slave 10 ---
        ms.handler.SlaveId = 10
        res10, err10 := ms.client.ReadHoldingRegisters(0, 2)

        if err10 == nil && len(res10) == 4 {
            r0_in := binary.BigEndian.Uint16(res10[0:2])
            r1_in := binary.BigEndian.Uint16(res10[2:4])
            currentData := Unpack32(r0_in, r1_in)

            // ShowResults(currentData, r0_in, r1_in)

            ms.state.UpdateSlave10(&currentData, true)

            // --- Синхронізація зі Slave 20 ---
            r0_out, r1_out := Pack32(&currentData)
            if r0_out != lastReg0 || r1_out != lastReg1 || firstRun {
                ms.handler.SlaveId = 20
                packedBytes := make([]byte, 4)
                binary.BigEndian.PutUint16(packedBytes[0:2], r0_out)
                binary.BigEndian.PutUint16(packedBytes[2:4], r1_out)

                time.Sleep(2 * time.Millisecond)
                _, err20 := ms.client.WriteMultipleRegisters(0, 2, packedBytes)
                if err20 == nil {
                    lastReg0, lastReg1 = r0_out, r1_out
                    firstRun = false
                    fmt.Printf("[%s] Slave 20 OK: Reg0=%04X, Reg1=%04X\n", 
                        time.Now().Format("15:04:05"), r0_out, r1_out)
                } else {
                    fmt.Printf("[%s] Slave 20 Error: %v\n", 
                        time.Now().Format("15:04:05"), err20)
                }
            }
        } else {
            if err10 != nil {
                fmt.Printf("[%s] Slave 10 Error: %v\n", time.Now().Format("15:04:05"), err10)
            }
            ms.state.UpdateSlave10(nil, false)
        }

        duration := time.Since(start).Milliseconds()
        ms.state.UpdateCycleTime(duration)
        time.Sleep(2 * time.Millisecond)
    }
}

func (ms *ModbusService) Close() {
    ms.handler.Close()
}

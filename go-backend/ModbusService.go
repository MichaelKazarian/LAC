package main

import (
    "encoding/binary"
    "fmt"
    "time"
    "github.com/goburrow/modbus"
)

type ModbusService struct {
  client  modbus.Client
  handler *modbus.RTUClientHandler
  //mu      sync.Mutex
}

func NewModbusService(device string, baud int) *ModbusService {
    handler := modbus.NewRTUClientHandler(device)
    handler.BaudRate = baud
    handler.DataBits = 8
    handler.Parity = "N"
    handler.StopBits = 1
    handler.Timeout = 200 * time.Millisecond
    
    if err := handler.Connect(); err != nil {
        panic(fmt.Sprintf("❌ Помилка порту: %v", err))
    }
    return &ModbusService{client: modbus.NewClient(handler), handler: handler}
}

// Read реалізує інтерфейс HardwareService
func (ms *ModbusService) Read() (uint16, [32]uint16, error) {
    sensor, err3 := ms.readEncoder()
    if err3 != nil {
        return 0, [32]uint16{}, err3
    }
    time.Sleep(2 * time.Millisecond)

    inputs, err10 := ms.readInputs()
    if err10 != nil {
        return sensor, [32]uint16{}, err10
    }
    time.Sleep(2 * time.Millisecond)
    return sensor, inputs, nil
}

func (ms *ModbusService) readEncoder() (uint16, error) {
    ms.handler.SlaveId = AddrEncoder
    res, err := ms.client.ReadHoldingRegisters(0, 1)
    if err != nil {
        return 0, fmt.Errorf("slave 3 error: %w", err)
    }
    return binary.BigEndian.Uint16(res), nil
}

func (ms *ModbusService) readInputs() ([32]uint16, error) {
    ms.handler.SlaveId = AddrInputs
    res, err := ms.client.ReadHoldingRegisters(0, 2)
    if err != nil {
        return [32]uint16{}, fmt.Errorf("slave 10 error: %w", err)
    }
    
    if len(res) != 4 {
        return [32]uint16{}, fmt.Errorf("slave 10: invalid data length")
    }

    r0 := binary.BigEndian.Uint16(res[0:2])
    r1 := binary.BigEndian.Uint16(res[2:4])
    
    return Unpack32(r0, r1), nil
}

// Write реалізує інтерфейс HardwareService
func (ms *ModbusService) Write(values [32]uint16) error {
    r0, r1 := Pack32(&values)
    packedBytes := make([]byte, 4)
    binary.BigEndian.PutUint16(packedBytes[0:2], r0)
    binary.BigEndian.PutUint16(packedBytes[2:4], r1)
    
    time.Sleep(2 * time.Millisecond)
    ms.handler.SlaveId = AddrOutputs
    _, err := ms.client.WriteMultipleRegisters(0, 2, packedBytes)
    ms.logWriteStatus(err, r0, r1)
    return err
}

func (ms *ModbusService) Close() error {
    return ms.handler.Close()
}

func (ms *ModbusService) logWriteStatus(err error, r0, r1 uint16) {
	timestamp := time.Now().Format("15:04:05")
	if err == nil {
		fmt.Printf("[%s] Slave 20 OK: Reg0=%04X, Reg1=%04X\n", timestamp, r0, r1)
	} else {
		fmt.Printf("[%s] Slave 20 Error: %v\n", timestamp, err)
	}
}

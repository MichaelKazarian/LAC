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
        panic(fmt.Sprintf("[MODBUS] Port error: %v", err))
    }
    return &ModbusService{client: modbus.NewClient(handler), handler: handler}
}

// Read реалізує інтерфейс HardwareService
func (ms *ModbusService) Read() (uint16, [32]uint16, error) {
  sensor, err := ms.readEncoder()
  if err != nil {
    return 0, [32]uint16{}, err
  }
  time.Sleep(2 * time.Millisecond)

  inputs, err := ms.readInputs()
  if err != nil {
    return sensor, [32]uint16{}, err
  }
  time.Sleep(2 * time.Millisecond)

  return sensor, inputs, nil
}

func (ms *ModbusService) readEncoder() (uint16, error) {
    ms.handler.SlaveId = AddrEncoder
    res, err := ms.client.ReadHoldingRegisters(0, 1)
    if err != nil {
        return 0, ms.logReadError("Encoder", AddrEncoder, err)
    }
    return binary.BigEndian.Uint16(res), nil
}

func (ms *ModbusService) readInputs() ([32]uint16, error) {
    ms.handler.SlaveId = AddrInputs
    res, err := ms.client.ReadHoldingRegisters(0, 2)
    if err != nil {
        return [32]uint16{}, ms.logReadError("Inputs Block", AddrInputs, err)
    }
    
    if len(res) != 4 {
        return [32]uint16{}, fmt.Errorf("[MODBUS] Inputs Block (addr %d): invalid data length", AddrInputs)
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

func (ms *ModbusService) logReadError(name string, addr byte, err error) error {
    return fmt.Errorf("[MODBUS] %s (addr %d) Read Error: %w", name, addr, err)
}

func (ms *ModbusService) logWriteStatus(err error, r0, r1 uint16) {
    timestamp := time.Now().Format("15:04:05")
    if err == nil {
        fmt.Printf("[%s] Outputs (addr %d) OK: [0x%04X 0x%04X]\n", 
            timestamp, AddrOutputs, r0, r1)
    } else {
        fmt.Printf("[%s] Outputs (addr %d) Write Error: %v\n", 
            timestamp, AddrOutputs, err)
    }
}

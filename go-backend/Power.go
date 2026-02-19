package main

// PowerControl — незалежний від Modbus канал керування живленням вихідної плати.
type PowerControl interface {
    EnableOutputsPower() error
    DisableOutputsPower() error
}

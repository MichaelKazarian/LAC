package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

const (
	DEVICE            = "/dev/ttyUSB0"
	BAUD_RATE         = 38400
	MODBUS_TIMEOUT    = 20 * time.Millisecond
	INTER_SLAVE_DELAY = 2 * time.Millisecond
)

func initModbus(device string, baud int, slaveID byte) (modbus.Client, *modbus.RTUClientHandler, error) {
	handler := modbus.NewRTUClientHandler(device)
	handler.BaudRate = baud
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = slaveID
	handler.Timeout = MODBUS_TIMEOUT

	// ❌ НЕ встановлюємо RS485 для CH341 - він не підтримує!
	// handler.RS485.Enabled = true
	
	err := handler.Connect()
	if err != nil {
		return nil, nil, err
	}

	return modbus.NewClient(handler), handler, nil
}

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║ Modbus Benchmark (CH341 USB-RS485)                    ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝\n")

	// Ініціалізація Modbus
	client, handler, err := initModbus(DEVICE, BAUD_RATE, 3)
	if err != nil {
		log.Fatalf("❌ Помилка Modbus: %v", err)
	}
	defer handler.Close()

	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║ Тестування швидкості                                  ║")
	fmt.Println("║ Baseline: перші 20 циклів                             ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝\n")

	// Статистика
	cycleCount := 0
	var totalTime int64
	var minCycle int64 = 9999
	var maxCycle int64 = 0
	var baseline int64

	type CycleStats struct {
		slave3Time  int64
		slave10Time int64
		totalTime   int64
	}
	var stats []CycleStats

	for {
		start := time.Now()

		// Slave 3
		handler.SlaveId = 3
		t1 := time.Now()
		res3, err3 := client.ReadHoldingRegisters(0, 1)
		d1 := time.Since(t1).Milliseconds()

		time.Sleep(INTER_SLAVE_DELAY)

		// Slave 10
		handler.SlaveId = 10
		t2 := time.Now()
		_, err10 := client.ReadHoldingRegisters(0, 2)
		d2 := time.Since(t2).Milliseconds()

		cycleTime := time.Since(start).Milliseconds()
		cycleCount++
		totalTime += cycleTime

		if cycleTime < minCycle {
			minCycle = cycleTime
		}
		if cycleTime > maxCycle {
			maxCycle = cycleTime
		}

		stats = append(stats, CycleStats{
			slave3Time:  d1,
			slave10Time: d2,
			totalTime:   cycleTime,
		})

		if len(stats) > 100 {
			stats = stats[1:]
		}

		if cycleCount == 20 {
			baseline = totalTime / 20
			fmt.Printf("\n📊 BASELINE (20 циклів):\n")
			fmt.Printf("   Середнє: %d мс\n", baseline)
			fmt.Printf("   Мінімум: %d мс\n", minCycle)
			fmt.Printf("   Максимум: %d мс\n\n", maxCycle)
		}

		if err3 == nil && err10 == nil {
			val3 := binary.BigEndian.Uint16(res3)
			avg := float64(totalTime) / float64(cycleCount)

			symbol := ""
			if cycleCount > 20 {
				if cycleTime < int64(avg)-2 {
					symbol = " 🚀"
				} else if cycleTime > int64(avg)+2 {
					symbol = " 🐌"
				}
			}

			fmt.Printf("[%04d] S3:%dms(%d°) S10:%dms | ⏱️ %dms%s | Avg:%.1f\n",
				cycleCount, d1, val3, d2, cycleTime, symbol, avg)
		} else {
			fmt.Printf("[%04d] ✗ Помилка: S3=%v S10=%v\n", cycleCount, err3, err10)
		}

		if cycleCount%50 == 0 && cycleCount > 0 {
			avg := float64(totalTime) / float64(cycleCount)

			var avgS3, avgS10 int64
			for _, s := range stats {
				avgS3 += s.slave3Time
				avgS10 += s.slave10Time
			}
			avgS3 /= int64(len(stats))
			avgS10 /= int64(len(stats))

			fmt.Printf("\n╔═════════════════════════════════════════════════════╗\n")
			fmt.Printf("║ 📊 СТАТИСТИКА (%d циклів)                          \n", cycleCount)
			fmt.Printf("╠═════════════════════════════════════════════════════╣\n")
			fmt.Printf("║ Повний цикл:                                       \n")
			fmt.Printf("║   Середнє:  %.1f мс                               \n", avg)
			fmt.Printf("║   Мінімум:  %d мс                                  \n", minCycle)
			fmt.Printf("║   Максимум: %d мс                                  \n", maxCycle)
			fmt.Printf("║   Розкид:   %d мс                                  \n", maxCycle-minCycle)
			fmt.Printf("║                                                    \n")
			fmt.Printf("║ По компонентах (avg останніх 100):                 \n")
			fmt.Printf("║   Slave 3:  %d мс                                  \n", avgS3)
			fmt.Printf("║   Slave 10: %d мс                                  \n", avgS10)
			fmt.Printf("║   Delays:   %d мс (2×%dms)                         \n", INTER_SLAVE_DELAY.Milliseconds()*2, INTER_SLAVE_DELAY.Milliseconds())

			if baseline > 0 {
				diff := baseline - int64(avg)
				percent := (float64(diff) / float64(baseline)) * 100
				fmt.Printf("║                                                    \n")
				fmt.Printf("║ Порівняння з baseline:                             \n")
				if diff > 0 {
					fmt.Printf("║   🚀 Швидше на %d мс (%.1f%%)                     \n", diff, percent)
				} else if diff < 0 {
					fmt.Printf("║   🐌 Повільніше на %d мс (%.1f%%)                 \n", -diff, -percent)
				} else {
					fmt.Printf("║   ⚖️  Ідентично baseline                           \n")
				}
			}

			fmt.Printf("╚═════════════════════════════════════════════════════╝\n\n")
		}

		time.Sleep(20 * time.Millisecond)
	}
}

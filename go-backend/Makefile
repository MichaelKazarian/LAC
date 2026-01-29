# Назва бінарного файлу
BINARY_NAME=modbus-app

# Параметри для Raspberry Pi (ARM)
TARGET_OS=linux
TARGET_ARCH=arm64

.PHONY: all init build build-arm run clean fmt deploy help

# 1. Початкова ініціалізація проекту
init:
	@if [ ! -f go.mod ]; then \
	go mod init modbus-app; \
	fi
	go get github.com/goburrow/modbus
	@echo "Проект ініціалізовано. Залежності завантажено."

# Стандартна команда: скомпілювати для поточної системи
build:
	go build -o bin/$(BINARY_NAME) .

# Крос-компіляція для Raspberry Pi 4
build-arm:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -o bin/$(BINARY_NAME)-arm .

# Запуск локально
run: build
	./bin/$(BINARY_NAME)

# Очищення
clean:
	go clean
	rm -rf bin/

# Форматування коду
fmt:
	go fmt ./...

# Відправка на Raspberry Pi (замініть pi@raspberrypi.local на ваш хост)
deploy: build-arm
	scp bin/$(BINARY_NAME)-arm pi@raspberrypi.local:/home/pi/$(BINARY_NAME)
	ssh pi@raspberrypi.local "chmod +x /home/pi/$(BINARY_NAME)"

# Допомога
help:
	@echo "Команди:"
	@echo "  make init       - створити go.mod та завантажити бібліотеки"
	@echo "  make build      - зібрати локально"
	@echo "  make build-arm  - зібрати для Raspberry Pi 4 (arm64)"
	@echo "  make run        - зібрати та запустити"
	@echo "  make deploy     - зібрати для ARM та відправити на RPi"
	@echo "  make clean      - видалити бінарники"

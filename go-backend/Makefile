HOST=localhost
PORT=8080
SERVICE_NAME=lac-backend.service
SERVICE_SOURCE=../deploy/$(SERVICE_NAME)
SYSTEMD_PATH=/etc/systemd/system/
DESKTOP_SOURCE=../deploy/TAK.desktop
DESKTOP_TARGET=$(HOME)/Desktop/TAK.desktop
AUTOSTART_DIR=$(HOME)/.config/autostart/
AUTOSTART_TARGET=$(AUTOSTART_DIR)TAK.desktop
UDEV_RULE_PATH=/etc/udev/rules.d/99-touchscreen.rules
TOUCHSCREEN_NAME=WaveShare WS170120

# Назва бінарного файлу
BINARY_NAME=modbus-app

# Параметри для Raspberry Pi (ARM)
TARGET_OS=linux
TARGET_ARCH=arm64

.PHONY: all init build build-arm run clean fmt deploy deploy-service deploy-lac-web-gui deploy-rotate-screen-90-degree deploy-rotate-screen-normal help

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
run: build fw-check
	cd bin && ./$(BINARY_NAME)

# Перевірка та налаштування iptables на віддаленій машині
fw-check:
	@echo "Перевірка прав на фаєрвол..."
	@if [ "$$(id -u)" = "0" ]; then \
		iptables -C INPUT -p tcp --dport $(PORT) -j ACCEPT > /dev/null 2>&1 || \
		iptables -I INPUT 1 -p tcp --dport $(PORT) -j ACCEPT; \
		echo "Фаєрвол налаштовано."; \
	else \
		echo "Запущено без sudo. Пропускаю налаштування фаєрволу."; \
		echo "Переконайтеся, що порт $(PORT) відкритий вручну."; \
	fi

# Очищення
clean:
	go clean
	rm -rf bin/

# Форматування коду
fmt:
	go fmt ./...

deploy-service:
	@echo "Встановлення системного сервісу..."
	@# Копіюємо файл сервісу з каталогу deploy
	cp $(SERVICE_SOURCE) $(SYSTEMD_PATH)
	@# Оновлюємо конфігурацію systemd
	systemctl daemon-reload
	@# Додаємо в автозавантаження та запускаємо
	systemctl enable $(SERVICE_NAME)
	systemctl start $(SERVICE_NAME)
	@echo "Сервіс $(SERVICE_NAME) успішно встановлено та запущено."

# deploy робила все разом
deploy: build fw-check deploy-service
	@echo "Повне розгортання завершено."

deploy-web-gui:
	@echo "--- Розгортання веб-інтерфейсу (Desktop + Autostart) ---"
	@if [ -f $(DESKTOP_SOURCE) ]; then \
		# 1. Копіюємо на робочий стіл \
		cp $(DESKTOP_SOURCE) $(DESKTOP_TARGET); \
		chmod +x $(DESKTOP_TARGET); \
		# 2. Створюємо папку автостарту та копіюємо туди \
		mkdir -p $(AUTOSTART_DIR); \
		cp $(DESKTOP_SOURCE) $(AUTOSTART_TARGET); \
		chmod +x $(AUTOSTART_TARGET); \
		echo "Веб-інтерфейс налаштовано на автозапуск та додано іконку."; \
	else \
		echo "Помилка: $(DESKTOP_SOURCE) не знайдено!"; \
		exit 1; \
	fi

# Поворот на 90 градусів (Landscape -> Portrait)
deploy-rotate-touch-90:
	@echo "--- Налаштування тачскріна: поворот 90 градусів ---"
	printf 'ACTION=="add", ATTRS{name}=="%s", ENV{LIBINPUT_CALIBRATION_MATRIX}="-0 -1 1 1 0 0 0 0 1"\n' "$(TOUCHSCREEN_NAME)" > $(UDEV_RULE_PATH)
	udevadm control --reload-rules && udevadm trigger
	@echo "Правила оновлено. Можливо, знадобиться перезапуск сесії браузера."

# Повернення в нормальний режим (закоментований рядок або пуста матриця)
deploy-rotate-touch-norm:
	@echo "--- Налаштування тачскріна: нормальний режим ---"
	printf '# ACTION=="add", ATTRS{name}=="%s", ENV{LIBINPUT_CALIBRATION_MATRIX}="-0 -1 1 1 0 0 0 0 1"\n' "$(TOUCHSCREEN_NAME)" > $(UDEV_RULE_PATH)
	udevadm control --reload-rules && udevadm trigger
	@echo "Правила повернуто до початкового стану."

help:
	@echo "Команди розробки:"
	@echo "  make init                - створити go.mod та завантажити бібліотеки"
	@echo "  make build               - зібрати локально"
	@echo "  make build-arm           - зібрати для Raspberry Pi 4 (arm64)"
	@echo "  make run                 - зібрати та запустити (локально)"
	@echo "  make clean               - видалити бінарники"
	@echo ""
	@echo "Команди деплою (потребують sudo):"
	@echo "  make deploy              - повний цикл: build + fw-check + сервіси"
	@echo "  make deploy-service      - встановити тільки go-backend сервіс"
	@echo "  make deploy-web-gui      - встановити Chromium Kiosk сервіс"
	@echo ""
	@echo "Налаштування nfxcrhsyf (потребують sudo):"
	@echo "  make deploy-rotate-touch-90      - повернути тач на 90 градусів"
	@echo "  make deploy-rotate-touch-norm  - повернути тач у стандартний режим"

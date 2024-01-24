upload:
	cd .. && platformio run --target upload

build:
	cd .. && (pio run) 2>&1 | awk '{if (/src\//){print substr($$0,5, 100)} else {print "",$$0}}'

clean:
	cd .. && platformio run --target clean

monitor:
	cd .. && platformio device monitor

init:
	pio project init --ide emacs --board pro16MHzatmega328


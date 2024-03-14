#include <Arduino.h>
#include <SimpleModbusSlave.h>

#define PRINT_SPEED 5 // Modbus update per second
#define BAUD_RATE 9600
#define BITS 64
void gpioInit();


const int MODBUS_ID = 10;
const int HOLDING_REGS_SIZE = 1;
unsigned int modbusHoldingRegisters[HOLDING_REGS_SIZE] = {0};

int RAWVALUES [BITS] = {};

void setup() {
  Serial.begin(BAUD_RATE);
  // modbus_configure(BAUD_RATE, MODBUS_ID, 1, HOLDING_REGS_SIZE, 0);
  gpioInit();
}

void loop() {
  // modbusHoldingRegisters[0] = degree;
  // printRaw(RAWVALUES); Serial.print("\t");
  // Serial.println( modbusHoldingRegisters[0] );
  // modbus_update(modbusHoldingRegisters);
  delay(PRINT_SPEED);
}

void gpioInit() {
  // for (int i = 0; i < BITS; i++) {
  //   pinMode(PINS[i], INPUT_PULLUP);
  // }
}


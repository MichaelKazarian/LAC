#include <Arduino.h>
#include <SimpleModbusSlave.h>
#include <MCP23017.h>

#define PRINT_SPEED 5 // Modbus update per second
#define BAUD_RATE 9600
#define BITS 64
#define MCP_LAST_PIN 6 // MCP23017 has problematic GPA7/GPB7
#define BOARD_IN_FIRST 7
#define BOARD_IN_LAST 10

MCP23017 mcp0 = MCP23017(0x20);

void gpioInit();
void mcpInit();
void registersInit();
void readPins();


const int MODBUS_ID = 10;
const int HOLDING_REGS_SIZE = 32;
unsigned int holdingRegisters[HOLDING_REGS_SIZE] = {};

int RAWVALUES [BITS] = {};

void setup() {
  // Serial.begin(BAUD_RATE);
  registersInit();
  gpioInit();
  mcpInit();
  modbus_configure(BAUD_RATE, MODBUS_ID, 1, HOLDING_REGS_SIZE, 0);
}

void loop() {
  readPins();
  modbus_update(holdingRegisters);
  delay(PRINT_SPEED);
}

void registersInit() {
  for (int i=0; i< HOLDING_REGS_SIZE; i++) holdingRegisters[i] = 0;
}

void gpioInit() {
  for (int i=BOARD_IN_FIRST; i<= BOARD_IN_LAST; i++) pinMode(i, INPUT_PULLUP);
}

/**
 * \brief Inits MCP23017 devices.
 */
void mcpInit() {
  // Init MCP23017 adressing pins
  pinMode(A1, OUTPUT); digitalWrite(A1, LOW);  // MCP0, addr 20
  pinMode(A2, OUTPUT); digitalWrite(A2, LOW);
  pinMode(A3, OUTPUT); digitalWrite(A3, LOW);
  mcp0.init();
  for (int i=0; i<= MCP_LAST_PIN; i++) { // GPA/GPB init
     mcp0.pinMode(i, INPUT_PULLUP); mcp0.pinMode(i+8, INPUT_PULLUP);
  }
}

void readPins() {
  for (int i=0; i<= MCP_LAST_PIN; i++) {
    holdingRegisters[i] = mcp0.digitalRead(i);       // Reading MCP0 GPA/GPB
    holdingRegisters[i+7] = mcp0.digitalRead(i+8);
  }

  int regNext = 28;
  for (int i=BOARD_IN_FIRST; i<= BOARD_IN_LAST; i++){
    holdingRegisters[regNext] = digitalRead(i);     // Reading Board pins
    regNext++;
  }
}

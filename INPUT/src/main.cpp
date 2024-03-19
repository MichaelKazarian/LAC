#include <Arduino.h>
#include <SimpleModbusSlave.h>
#include <MCP23017.h>

#define PRINT_SPEED 5 // Modbus update per second
#define BAUD_RATE 9600
#define BITS 64
#define MCP_LAST_PIN 6 // MCP23017 has problematic GPA7/GPB7
#define BOARD_IN_FIRST 7
#define BOARD_IN_LAST 10
#define GPA_SHIFT_B 8

MCP23017 mcp0 = MCP23017(0x20);
MCP23017 mcp1 = MCP23017(0x21);

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
  modbus_update(holdingRegisters);
  readPins();
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
  pinMode(2, OUTPUT);  digitalWrite(2, LOW);   // MCP0, addr 20
  pinMode(3, OUTPUT);  digitalWrite(3, LOW);
  pinMode(4, OUTPUT);  digitalWrite(4, LOW);
  pinMode(A1, OUTPUT); digitalWrite(A1, LOW);  // MCP1, addr 21
  pinMode(A2, OUTPUT); digitalWrite(A2, LOW);
  pinMode(A3, OUTPUT); digitalWrite(A3, HIGH);
  mcp0.init();
  mcp1.init();
  for (int i=0; i<= MCP_LAST_PIN; i++) { // GPA/GPB init
     mcp0.pinMode(i, INPUT_PULLUP); mcp0.pinMode(i+GPA_SHIFT_B, INPUT_PULLUP);
     mcp1.pinMode(i, INPUT_PULLUP); mcp1.pinMode(i+GPA_SHIFT_B, INPUT_PULLUP);
  }
}

void readPins() {
  void fixBoardPinOrder(int pin0, int pin1);
  const char PRC_PINS = 4;   // Processor pins are in the middle of the board
  const char DS       = 14;  // Direction shift
  const char HS       = 7;   // GPA/GBP shift in holding registers
  // Write holding registers in reverce order according to board design
  for (int i=0; i<= MCP_LAST_PIN; i++) {
    holdingRegisters[MCP_LAST_PIN-i+HS] = mcp1.digitalRead(i); // GBP
    holdingRegisters[MCP_LAST_PIN-i] = mcp1.digitalRead(i+GPA_SHIFT_B); // GPA
    holdingRegisters[MCP_LAST_PIN-i+DS+HS+PRC_PINS] = mcp0.digitalRead(i); // GPA
    holdingRegisters[MCP_LAST_PIN-i+DS+PRC_PINS] = mcp0.digitalRead(i+GPA_SHIFT_B); //GPB
  }

  int regBrdPins = 17; // Reading processor pins (PRC_PINS)
  for (int i=BOARD_IN_FIRST; i<= BOARD_IN_LAST; i++) {
    holdingRegisters[regBrdPins--] = digitalRead(i);
  }
  fixBoardPinOrder(4, 6); fixBoardPinOrder(22, 24);
}

void fixBoardPinOrder(int pin0, int pin1) {
  char tmp = holdingRegisters[pin0];
  holdingRegisters[pin0] = holdingRegisters[pin1];
  holdingRegisters[pin1] = tmp;
}

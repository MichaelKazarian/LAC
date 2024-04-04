#include <Arduino.h>
#include <SimpleModbusSlave.h>
#include <MCP23017.h>

#define PRINT_SPEED 5 // Modbus update per second
#define BAUD_RATE 9600
#define MCP_LAST_PIN 6 // MCP23017 has problematic GPA7/GPB7
#define BOARD_IN_FIRST 7
#define BOARD_IN_LAST 10
#define GPA_SHIFT_B 8

MCP23017 mcp0 = MCP23017(0x20);
MCP23017 mcp1 = MCP23017(0x21);

void gpioInit();
void mcpInit();
void registersInit();
void writePins();


const int MODBUS_ID = 10;
const int HOLDING_REGS_SIZE = 32;
unsigned int holdingRegisters[HOLDING_REGS_SIZE] = {};

void setup() {
  while (!Serial) { delay(10); }
  Serial.begin(BAUD_RATE);
  registersInit();
  gpioInit();
  mcpInit();
  modbus_configure(BAUD_RATE, MODBUS_ID, 13, HOLDING_REGS_SIZE, 0);
}

void loop() {
  modbus_update(holdingRegisters);
  writePins();
}

void registersInit() {
  for (int i=0; i< HOLDING_REGS_SIZE; i++) holdingRegisters[i] = LOW;
}

void gpioInit() {
  for (int i=BOARD_IN_FIRST; i<= BOARD_IN_LAST; i++) pinMode(i, OUTPUT);
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
  pinMode(A3, OUTPUT); digitalWrite(A3, LOW);
  mcp0.init();
  // mcp1.init();
  for (int i=0; i<= MCP_LAST_PIN; i++) { // GPA/GPB init
    mcp0.pinMode(i, OUTPUT); mcp0.pinMode(i+GPA_SHIFT_B, OUTPUT);
    // mcp1.pinMode(i, OUTPUT); mcp1.pinMode(i+GPA_SHIFT_B, OUTPUT);
  }
}

void writePins() {
  void fixBoardPinOrder(int pin0, int pin1);
  fixBoardPinOrder(4, 6); fixBoardPinOrder(22, 24);
  const char PRC_PINS = 4;   // Processor pins are in the middle of the board
  const char DS       = 14;  // Direction shift
  const char HS       = 7;   // GPA/GBP shift in holding registers
  // Write holding registers in reverce order according to board design
  for (int i=0; i<= MCP_LAST_PIN; i++) {
    // mcp1.digitalWrite(i, !holdingRegisters[MCP_LAST_PIN-i+HS]);          // GBP
    // mcp1.digitalWrite(i+GPA_SHIFT_B, !holdingRegisters[MCP_LAST_PIN-i]); // GPA
    mcp0.digitalWrite(i, !holdingRegisters[MCP_LAST_PIN-i+DS+HS+PRC_PINS]); // GPA
    mcp0.digitalWrite(i+GPA_SHIFT_B, !holdingRegisters[MCP_LAST_PIN-i+DS+PRC_PINS]); //GPB
  }

  int regBrdPins = 17; // Reading processor pins (PRC_PINS)
  for (int i=BOARD_IN_FIRST; i<= BOARD_IN_LAST; i++) {
    digitalWrite(i, !holdingRegisters[regBrdPins--]);
  }
}

void fixBoardPinOrder(int pin0, int pin1) {
  char tmp = holdingRegisters[pin0];
  holdingRegisters[pin0] = holdingRegisters[pin1];
  holdingRegisters[pin1] = tmp;
}

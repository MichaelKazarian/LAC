#include <Arduino.h>
#include <SimpleModbusSlave.h>
#include <MCP23017.h>

#define PRINT_SPEED 5 // Modbus update per second
#define BAUD_RATE 9600
#define MCP_LAST_PIN 2 // MCP23017 has problematic GPA7/GPB7
#define MCP_OUTPUT_SHIFT 4

MCP23017 mcp = MCP23017(0x20);

void gpioInit();
void mcpInit();
void registersInit();
void testMcp();


const int MODBUS_ID = 20;
const int HOLDING_REGS_SIZE = 3;
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
  testMcp();
}

void registersInit() {
  for (int i=0; i< HOLDING_REGS_SIZE; i++) holdingRegisters[i] = LOW;
}

void gpioInit() {
  // for (int i=BOARD_IN_FIRST; i<= BOARD_IN_LAST; i++) pinMode(i, OUTPUT);
}

/**
 * \brief Inits MCP23017 devices.
 */
void mcpInit() {
  // Init MCP23017 adressing pins
  pinMode(2, OUTPUT);  digitalWrite(2, LOW);   // MCP0, addr 20
  pinMode(3, OUTPUT);  digitalWrite(3, LOW);
  pinMode(4, OUTPUT);  digitalWrite(4, LOW);
  mcp.init();
  for (int i=0; i<= MCP_LAST_PIN; i++) {
    mcp.pinMode(i, INPUT_PULLUP);             // pins 0-2
    mcp.pinMode(i+MCP_OUTPUT_SHIFT, OUTPUT);  // pins 4-6
  }
}

void testMcp() {
  for (int i=0; i<= MCP_LAST_PIN; i++) {
    holdingRegisters[i] = mcp.digitalRead(i);
    mcp.digitalWrite(i+MCP_OUTPUT_SHIFT, holdingRegisters[i]); // Input -> output transfer
  }
}

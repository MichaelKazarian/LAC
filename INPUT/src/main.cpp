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
void readPins();


const int MODBUS_ID = 10;
const int HOLDING_REGS_SIZE = 32;
unsigned int holdingRegisters[HOLDING_REGS_SIZE] = {};

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
  holdingRegisters[0] = !mcp1.digitalRead(14);
  holdingRegisters[1] = !mcp1.digitalRead(13);
  holdingRegisters[2] = !mcp1.digitalRead(12);
  holdingRegisters[3] = !mcp1.digitalRead(11);
  holdingRegisters[4] = !mcp1.digitalRead(8);
  holdingRegisters[5] = !mcp1.digitalRead(9);
  holdingRegisters[6] = !mcp1.digitalRead(10);
  holdingRegisters[7] = !mcp1.digitalRead(6);
  holdingRegisters[8] = !mcp1.digitalRead(5);
  holdingRegisters[9] = !mcp1.digitalRead(4);
  holdingRegisters[10] = !mcp1.digitalRead(3);
  holdingRegisters[11] = !mcp1.digitalRead(2);
  holdingRegisters[12] = !mcp1.digitalRead(1);
  holdingRegisters[13] = !mcp1.digitalRead(0);
  holdingRegisters[14] = !digitalRead(10);
  holdingRegisters[15] = !digitalRead(9);
  holdingRegisters[16] = !digitalRead(8);
  holdingRegisters[17] = !digitalRead(7);
  holdingRegisters[18] = !mcp0.digitalRead(14);
  holdingRegisters[19] = !mcp0.digitalRead(13);
  holdingRegisters[20] = !mcp0.digitalRead(12);
  holdingRegisters[21] = !mcp0.digitalRead(11);
  holdingRegisters[22] = !mcp0.digitalRead(8);
  holdingRegisters[23] = !mcp0.digitalRead(9);
  holdingRegisters[24] = !mcp0.digitalRead(10);
  holdingRegisters[25] = !mcp0.digitalRead(6);
  holdingRegisters[26] = !mcp0.digitalRead(5);
  holdingRegisters[27] = !mcp0.digitalRead(4);
  holdingRegisters[28] = !mcp0.digitalRead(3);
  holdingRegisters[29] = !mcp0.digitalRead(2);
  holdingRegisters[30] = !mcp0.digitalRead(1);
  holdingRegisters[31] = !mcp0.digitalRead(0);
}

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
  pinMode(A3, OUTPUT); digitalWrite(A3, HIGH);
  mcp0.init();
  mcp1.init();
  for (int i=0; i<= MCP_LAST_PIN; i++) {       // GPA/GPB init
    mcp0.pinMode(i, OUTPUT); mcp0.pinMode(i+GPA_SHIFT_B, OUTPUT);
    mcp1.pinMode(i, OUTPUT); mcp1.pinMode(i+GPA_SHIFT_B, OUTPUT);
    mcp0.digitalWrite(i, LOW); mcp0.digitalWrite(i+GPA_SHIFT_B, LOW);
    mcp1.digitalWrite(i, LOW); mcp1.digitalWrite(i+GPA_SHIFT_B, LOW);
  }
}

void writePins() {
  mcp1.digitalWrite(14, holdingRegisters[0]);
  mcp1.digitalWrite(13, holdingRegisters[1]);
  mcp1.digitalWrite(12, holdingRegisters[2]);
  mcp1.digitalWrite(11, holdingRegisters[3]);
  mcp1.digitalWrite(8, holdingRegisters[4]);
  mcp1.digitalWrite(9, holdingRegisters[5]);
  mcp1.digitalWrite(10, holdingRegisters[6]);
  mcp1.digitalWrite(6, holdingRegisters[7]);
  mcp1.digitalWrite(5, holdingRegisters[8]);
  mcp1.digitalWrite(4, holdingRegisters[9]);
  mcp1.digitalWrite(3, holdingRegisters[10]);
  mcp1.digitalWrite(2, holdingRegisters[11]);
  mcp1.digitalWrite(1, holdingRegisters[12]);
  mcp1.digitalWrite(0, holdingRegisters[13]);
  digitalWrite(10, holdingRegisters[14]);
  digitalWrite(9, holdingRegisters[15]);
  digitalWrite(8, holdingRegisters[16]);
  digitalWrite(7, holdingRegisters[17]);
  mcp0.digitalWrite(14, holdingRegisters[18]);
  mcp0.digitalWrite(13, holdingRegisters[19]);
  mcp0.digitalWrite(12, holdingRegisters[20]);
  mcp0.digitalWrite(11, holdingRegisters[21]);
  mcp0.digitalWrite(8, holdingRegisters[22]);
  mcp0.digitalWrite(9, holdingRegisters[23]);
  mcp0.digitalWrite(10, holdingRegisters[24]);
  mcp0.digitalWrite(6, holdingRegisters[25]);
  mcp0.digitalWrite(5, holdingRegisters[26]);
  mcp0.digitalWrite(4, holdingRegisters[27]);
  mcp0.digitalWrite(3, holdingRegisters[28]);
  mcp0.digitalWrite(2, holdingRegisters[29]);
  mcp0.digitalWrite(1, holdingRegisters[30]);
  mcp0.digitalWrite(0, holdingRegisters[31]);
}

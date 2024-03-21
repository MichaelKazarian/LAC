/**
 * Sample of use.
 * mbpoll -m rtu -a 10 -b 9600 -t 4 -r 1 -P none /dev/ttyUSB0 1 0 1
 * It works with an external 485 to TTL adapter.
 * Arduino nano internal USB adapter crashes with Connection timeout.
 * Also 2-3 connection timeout caught while mbpoll starts reading:
 * mbpoll -m rtu -a 10 -b 9600 -t 4 -r 1 -P none /dev/ttyUSB0
 * ....
 * Data type.............: 16-bit register, output (holding) register table
 *
 * -- Polling slave 10... Ctrl-C to stop)
 * Read output (holding) register failed: Connection timed out
 * -- Polling slave 10... Ctrl-C to stop)
 * [1]:    0
 * -- Polling slave 10... Ctrl-C to stop)
 * [1]:    0
 * ^C--- /dev/ttyUSB0 poll statistics ---
 * 3 frames transmitted, 2 received, 1 errors, 33.3% frame loss
 * TODO: Possible DTR interference, try to run using USB to TTL
 * adapter with RX/TX only
 */

#include <Arduino.h>
#include <SimpleModbusSlave.h>
#include <MCP23017.h>

#define PRINT_SPEED 5 // Modbus update per second
#define BAUD_RATE 9600
#define MCP_LAST_PIN 6 // MCP23017 has problematic GPA7/GPB7
#define BOARD_IN_FIRST 7
#define BOARD_IN_LAST 10

MCP23017 mcp0 = MCP23017(0x20);

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
  // delay(PRINT_SPEED);
}

void registersInit() {
  for (int i=0; i< HOLDING_REGS_SIZE; i++) holdingRegisters[i] = 0;
}

void gpioInit() {
  for (int i=BOARD_IN_FIRST; i<= BOARD_IN_LAST; i++) pinMode(i, OUTPUT);
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
     mcp0.pinMode(i, OUTPUT); mcp0.pinMode(i+8, OUTPUT);
  }
}

void writePins() {
  mcp0.digitalWrite(4, holdingRegisters[0]);
  mcp0.digitalWrite(5, holdingRegisters[1]);
  mcp0.digitalWrite(6, holdingRegisters[2]);
}

#include <Arduino.h>
#include <SimpleModbusSlave.h>

#define BITS 11  ///< Encoder data bits length
#define PRINT_SPEED 50 // Per second
#define BAUD_RATE 9600

void gpioInit();
void readRaw(int *a);
void printRaw(int *a);
int bcd8421(int *a, int length);
int rawToBin(int *a);

const int MODBUS_ID = 3;
const int HOLDING_REGS_SIZE = 1;
unsigned int modbusHoldingRegisters[HOLDING_REGS_SIZE] = {0};

const byte PINS [BITS] = { 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12 };
int RAWVALUES [BITS] = {};

void setup() {
  // Serial.begin(BAUD_RATE);
  modbus_configure(BAUD_RATE, MODBUS_ID, 1, HOLDING_REGS_SIZE, 0);
  gpioInit();
}

void loop() {
  readRaw(RAWVALUES);
  int degree = rawToBin(RAWVALUES);
  modbusHoldingRegisters[0] = degree;
  // printRaw(RAWVALUES); Serial.print("\t");
  // Serial.println( modbusHoldingRegisters[0] );
  modbus_update(modbusHoldingRegisters);
  delay(PRINT_SPEED);
}

void gpioInit() {
  for (int i = 0; i < BITS; i++) {
    pinMode(PINS[i], INPUT_PULLUP);
  }
}

/**
 * \brief Converts BCD 8421 code to a decimal value.
 *
 * \param[*a]      the pointer to the binary data array. Data order: 1 2 4 8.
 * \param[length]  data length, max value is 4.
 * \returns        0-9 value.
 */
int bcd8421(int *a, int length) {
  const int weight[] = {1, 2, 4, 8};
  int sum = 0;
  for (int i=0; i<length; ++i)
    sum += a[i]*weight[i];
  return sum;
};

/**
 * \brief Reads encoder binarry data to \param[*a]
 */
void readRaw(int *a) {
  int readVal;
  for (int i = 0; i < BITS; i++) {
    readVal = !digitalRead(PINS[i]);
    a[i] = readVal;
  }
}

/**
 * \brief Pretty print \param[*a] binary data without ending \n
 */
void printRaw(int *a) {
  for(int j = BITS-1; j >=0; j--) {
    if (j==7 || j == 3) {
      Serial.print(" ");
    };
    Serial.print(a[j]);
  }
}


/**
 * \brief Converts raw alue to decimal value.
 *
 * \param[*a]    the pointer to the binary data array.
 * \returns      value in the range of 0-719 .
 */
int rawToBin(int *a) {
  int x1   = bcd8421(a, 4);
  int x10  = bcd8421(&a[4], 4) * 10;
  int x100 = bcd8421(&a[8], 3) * 100;
  return x1+x10+x100;
}

/**
 * \brief Grey code test implementation.
 */
void greytobinary(int *a, int count) {
  for(int i = count - 2; i >= 0; i--) {
    a[i] = a[i]^a[i+1];
  }
}

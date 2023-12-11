#include <Arduino.h>
#define BITS 11 ///< Encoder data bits length
#define MSB 12  ///< The GPIO 12 is the most significant bit on the board.
                ///< The least significant bit is GPIO 2.

void gpioInit();
void readRaw(int *a);
void printRaw(int *a);
int bcd8421(int *a, int length);
int rawToBin(int *a);

void setup() {
  Serial.begin(9600);
  gpioInit();
}

void loop() {
  int rawValues[BITS];
  readRaw(rawValues);
  printRaw(rawValues);
  Serial.print("\t");
  Serial.println( rawToBin(rawValues) );
  Serial.print("\n");
  delay(100);
}

void gpioInit() {
  int i;
  for (i = MSB; i > MSB - BITS; i--){
    pinMode(i, INPUT_PULLUP);
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
  int i, j=BITS-1;
  for (i = MSB; i > MSB-BITS; i--) {
    int readVal;
    readVal = !digitalRead(i);
    a[j] = readVal;
    j--;
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


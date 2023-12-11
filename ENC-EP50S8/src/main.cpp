#include <Arduino.h>
#define BITS 11
#define LSB 12

void gpioInit();
void readRaw(int *a);
void printRaw(int *a);
int bcd8421(int *a, int length, int factor);
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
  for (i = LSB; i > LSB - BITS; i--){
    pinMode(i, INPUT_PULLUP);
  }
}

int bcd8421(int *a, int length, int factor){
  const int weight[] = {1, 2, 4, 8};
  int sum = 0;
  for (int i=0; i<length; ++i)
    sum += a[i]*weight[i];
  return sum*factor;
};

void readRaw(int *a) {
  int i, j=BITS-1;
  for (i = LSB; i > LSB-BITS; i--){ 
    int readVal;
    readVal = !digitalRead(i);
    a[j] = readVal;
    j--;
  }
}

void printRaw(int *a) {
  for(int j = BITS-1; j >=0; j--){
    if (j==7 || j == 3) {
      Serial.print(" ");
    };
    Serial.print(a[j]);
  }
}

int rawToBin(int *a) {
  int x1   = bcd8421(a, 4 ,1);
  int x10  = bcd8421(&a[4], 4 ,10);
  int x100 = bcd8421(&a[8], 3 ,100);
  return x1+x10+x100;
}

void greytobinary(int *a, int count) {
  int i;
  for(i = count - 2; i >= 0; i--) {
    a[i] = a[i]^a[i+1];
  }
}


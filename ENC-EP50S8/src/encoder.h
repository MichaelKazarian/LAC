/**
 * EP50S8-720-1F-1-24 pin constants
 *
 * The encoder output has a weight based on 2, an exponent of N,
 * and is a BCD code of 8-4-2-1.
 */
#define BITS          11 ///< Encoder data bits length

#define OUT2E0        16 ///< PC2, D16, Val is 1, BROWN
#define OUT2E1        17 ///< PC3, D17, Val IS 2, RED
#define OUT2E2        18 ///< PC4, D18, Val is 4, ORANGE
#define OUT2E3        19 ///< PC5, D19, Val is 8, YELLOW

#define OUT2E0x10     12 ///< PB4, D12, Val is 10, BLUE
#define OUT2E1x10     14 ///< PC0, D14, Val is 20, VIOLET
#define OUT2E2x10     15 ///< PC1, D15, Val is 40, GRAY
#define OUT2E3x10      8 ///< PB0, D8,  Val is 80, WHITE/BROWN

#define OUT2E0x100     9 ///< PB1, D9,  Val is 100, WHITE/RED
#define OUT2E1x100    10 ///< PB2, D10, Val is 200, WHITE/ORANGE
#define OUT2E2x100    11 ///< PB3, D11, Val is 400, WHITE/YELLOW

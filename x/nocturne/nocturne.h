#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

/**
 * Minimum entropy threshold for a valid trauma event.
 */
#define MIN_ENTROPY 2.5

/**
 * Maximum allowable suffering score for a dream to be accepted.
 */
#define MAX_SUFFERING 0.05

#define RECOGNITION_THRESHOLD 0.8

/**
 * Threshold for cosmic coherence (ΔI) during mirroring.
 */
#define COHERENCE_THRESHOLD 0.7

char *hello_nocturne(void);

void nocturne_free_string(char *s);

char *simulate_qlink(void);

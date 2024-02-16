#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct ModuleDescriptor {
  const uint8_t *name;
  uintptr_t name_len;
} ModuleDescriptor;

typedef struct InitData {
  const uint8_t *proto_file_descriptors;
  uintptr_t proto_file_descriptors_len;
  const struct ModuleDescriptor *module_descriptors;
  uintptr_t num_modules;
} InitData;

const struct InitData *__init(void);

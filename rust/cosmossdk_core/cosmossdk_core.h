#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

typedef uint32_t (*UnaryMethodHandler)(uint32_t ctx,
                                       const uint8_t *req,
                                       uintptr_t req_len,
                                       uint8_t *res,
                                       uintptr_t *res_len);

typedef struct ModuleInitData {
  const uint8_t *config;
  uint32_t config_len;
  uint32_t (*register_unary_method)(const uint8_t *service,
                                    uintptr_t service_len,
                                    const uint8_t *method,
                                    uintptr_t method_len,
                                    UnaryMethodHandler handler);
} ModuleInitData;

typedef int32_t (*ModuleInitFn)(const struct ModuleInitData *init_data);

typedef struct InitData {
  const uint8_t *proto_file_descriptors;
  uintptr_t proto_file_descriptors_len;
  uintptr_t num_modules;
  const uint8_t *const *module_names;
  const ModuleInitFn *module_init_fns;
} InitData;

const struct InitData *__init(void);

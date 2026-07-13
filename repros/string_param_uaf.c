/* Minimal AddressSanitizer repro for plcs_eval_ctx_set_str_eval_param
 * ownership. */
#include <dd/policies/eval_ctx.h>

#include <stdlib.h>
#include <string.h>

int main(void) {
  if (plcs_eval_ctx_init() != PLCS_ESUCCESS) {
    return 2;
  }

  char *value = strdup("FOO=bar");
  if (value == NULL) {
    return 2;
  }

  if (plcs_eval_ctx_set_str_eval_param(PLCS_STR_EVAL_PROCESS_ENVAR, value) !=
      PLCS_ESUCCESS) {
    free(value);
    return 2;
  }
  free(value);

  /* The current implementation returns the freed allocation. */
  return strcmp(plcs_eval_ctx_get_string_param(PLCS_STR_EVAL_PROCESS_ENVAR),
                "FOO=bar");
}

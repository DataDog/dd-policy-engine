# Confirmed bug repros

The repro tests are opt-in because they assert the intended behavior and fail
while the corresponding bugs are present.

## C runtime

Configure a test build with `DD_POLICY_BUILD_REPROS=ON`, build it, then run:

```sh
ctest --test-dir <build-dir> --output-on-failure -R '^repro_'
```

The string-parameter ownership repro requires AddressSanitizer to fail
reliably. The `clang-dev` preset enables it.

On platforms where that preset is unavailable, the ownership issue also has a
standalone repro:

```sh
clang -g -fsanitize=address \
  -I c/include -I c/src -I c/src/schema -I <flatcc>/include \
  c/src/eval_ctx.c repros/string_param_uaf.c -o /tmp/string-param-uaf
/tmp/string-param-uaf
```

## Go requirements converter

From `go/`, run:

```sh
go test -tags=repro ./repros
```

The two tests demonstrate an incorrect libc patch-version match and silent
acceptance of a deny rule whose only condition is an unsupported OS.

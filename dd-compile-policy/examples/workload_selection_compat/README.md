# Workload Selection Compatibility Examples

These examples document how old `dd-compile-policy` binaries handle JSON shapes
that may be useful for future Kubernetes workload selection and tracer injection
configuration.

The Agent passes APM workload selection Remote Config payloads to
`dd-compile-policy --input-string ...`. The compiler is strict about object
fields, but older compiler builds fall back to enum value `0` for unknown enum
names.

## Future Enums

`future_action_and_k8s_selectors.json` shows a possible future policy shape:

- `INJECT_TRACERS` action with `JAVA:1.42.0` in `values`.
- Kubernetes tag/label evaluators using `In`, `NotIn`, and `Exists`-style
  matching.

Current old compilers accept this JSON, but encode the unknown enum names as
their `*_UNKNOWN` values:

- `INJECT_TRACERS` becomes `ACTION_UNKNOWN`.
- `K8S_POD_LABEL` and `K8S_NAMESPACE_LABEL` become `STRING_EVAL_UNKNOWN`.
- `CMP_K8S_IN`, `CMP_K8S_NOT_IN`, and `CMP_K8S_EXISTS` become
  `CMP_STR_UNKNOWN`.

The string payloads, such as `JAVA:1.42.0` and `app=checkout`, are preserved in
the buffer. On old runtimes, the resulting action is a no-op and the future
selectors evaluate as abstain.

## Opaque Metadata

`opaque_target_metadata.json` shows target metadata encoded as strings in
existing schema fields. Old compilers accept this because the data is just a
string in `description` and `actions.values`.

This is only transport-compatible. A runtime or action handler must explicitly
parse the string payload for it to affect behavior.

## Rejected Fields

Do not place the `Target`-style fields from the Cluster Agent
autoinstrumentation config directly on a `Policy`. Old compilers reject that
shape with `unknown field`, because fields such as `name`, `podSelector`,
`namespaceSelector`, `ddTraceVersions`, and `ddTraceConfigs` are not fields in
the FlatBuffers `Policy` table.

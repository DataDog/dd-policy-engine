// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache 2.0 License. This product includes software developed at
// Datadog (https://www.datadoghq.com/).
//
// Copyright 2025-Present Datadog, Inc.

//go:build conformance_cgo

// This file is the cgo bridge to the real C engine (libpolicies.a), used by the
// cross-engine conformance test (cross_test.go). cgo is not allowed in _test.go
// files, so the bridge lives here in a regular, build-tagged source file and the
// test calls into it. It is deliberately kept in this separate "conformance"
// package (not "policies") so that its FlatBuffers/schema imports never become
// dependencies of the importable "policies" package.
//
// It is gated behind the "conformance_cgo" build tag and requires CGO_ENABLED=1
// plus the C toolchain, so the default CGO_ENABLED=0 builds and the portable
// Go-only corpus test (corpus_test.go) are unaffected. Build/run with:
//
//	make -C go conformance-cross
//
// The C engine is generic: label and ALWAYS_* semantics are not baked into the
// engine but provided by the host. The cgo preamble below registers the host
// evaluators that mirror exactly what the Go engine bakes in (constant ALWAYS_*
// leaves, a "key=value"/"key=" label evaluator, and numeric evaluators that
// abstain on an unset source), so the two engines are compared on identical
// semantics rather than on accidental host wiring.

package conformance

/*
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include <dd/policies/error_codes.h>
#include <dd/policies/eval_ctx.h>
#include <dd/policies/evaluator_default.h>
#include <dd/policies/evaluation_result.h>
#include <dd/policies/evaluator_types.h>

#include "nodes_reader.h"

// evaluate_rules is internal to the engine (not part of the public API) but is
// exported from the static library; declaring it here lets the harness drive
// the tree-walk directly and read back the tri-state result.
extern plcs_evaluation_result evaluate_rules(dd_wls_NodeTypeWrapper_table_t node, int depth);

#define CONF_MAX_LABELS 32
#define CONF_STR_MAX 256

// conf_label_set mirrors the Go Context.Labels[id] map: a per-id set of
// key->value pairs plus a "present" flag distinguishing "source unavailable"
// (ABSTAIN) from "source present but key absent" (FALSE).
typedef struct {
  int present;
  int count;
  char keys[CONF_MAX_LABELS][CONF_STR_MAX];
  char vals[CONF_MAX_LABELS][CONF_STR_MAX];
} conf_label_set;

static conf_label_set conf_labels[PLCS_STR_EVAL__COUNT];
static char conf_str_params[PLCS_STR_EVAL__COUNT][CONF_STR_MAX];

static void conf_copy(char *dst, const char *src) {
  if (!src) {
    dst[0] = '\0';
    return;
  }
  strncpy(dst, src, CONF_STR_MAX - 1);
  dst[CONF_STR_MAX - 1] = '\0';
}

// Constant evaluators mirror the Go ALWAYS_* leaves (evalString short-circuit).
static plcs_evaluation_result
conf_always_true(const char *p, const plcs_string_comparator c, const char *x, const char *d, plcs_string_evaluators id) {
  (void)p; (void)c; (void)x; (void)d; (void)id;
  return PLCS_EVAL_RESULT_TRUE;
}

static plcs_evaluation_result
conf_always_false(const char *p, const plcs_string_comparator c, const char *x, const char *d, plcs_string_evaluators id) {
  (void)p; (void)c; (void)x; (void)d; (void)id;
  return PLCS_EVAL_RESULT_FALSE;
}

static plcs_evaluation_result
conf_always_abstain(const char *p, const plcs_string_comparator c, const char *x, const char *d, plcs_string_evaluators id) {
  (void)p; (void)c; (void)x; (void)d; (void)id;
  return PLCS_EVAL_RESULT_ABSTAIN;
}

// conf_num / conf_unum mirror the Go numeric leaves: a missing fact (the
// NOT_SET sentinel) is ABSTAIN, not a comparison against the sentinel. This is
// the one place the default C evaluator differs from the Go model, so the host
// supplies the abstaining behavior rather than mutating the engine.
static plcs_evaluation_result
conf_num(const long policy, const plcs_numeric_comparator cmp, const long ctx, const char *d, plcs_numeric_evaluators id) {
  if (ctx == PLCS_NUM_NOT_SET) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }
  return plcs_default_numeric_evaluator(policy, cmp, ctx, d, id);
}

static plcs_evaluation_result
conf_unum(const unsigned long policy, const plcs_numeric_comparator cmp, const unsigned long ctx, const char *d, plcs_numeric_evaluators id) {
  if (ctx == PLCS_UNUM_NOT_SET) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }
  return plcs_default_unumeric_evaluator(policy, cmp, ctx, d, id);
}

// conf_label replicates the Go label semantics (model.go labelResult +
// dd_wls.go decodeLabel): the policy value is "key=value" (or "key=" with a
// prefix comparison for an existence check), resolved against the per-id label
// map. An unavailable source abstains; an absent key is false; otherwise the
// value is compared with the same comparator semantics as the default string
// evaluator.
static plcs_evaluation_result
conf_label(const char *policy, const plcs_string_comparator cmp, const char *ctx, const char *d, plcs_string_evaluators id) {
  (void)ctx;
  if (!policy || id < 0 || id >= PLCS_STR_EVAL__COUNT) {
    return PLCS_EVAL_RESULT_ABSTAIN;
  }
  conf_label_set *set = &conf_labels[id];
  if (!set->present) {
    return PLCS_EVAL_RESULT_ABSTAIN; // label source unavailable in this environment
  }
  const char *eq = strchr(policy, '=');
  if (!eq) {
    return PLCS_EVAL_RESULT_ABSTAIN; // malformed; the Go decoder rejects this at parse time
  }
  size_t klen = (size_t)(eq - policy);
  const char *val = eq + 1;

  const char *got = NULL;
  for (int i = 0; i < set->count; ++i) {
    if (strlen(set->keys[i]) == klen && strncmp(set->keys[i], policy, klen) == 0) {
      got = set->vals[i];
      break;
    }
  }

  // "key=" + CMP_PREFIX is the existence convention (Go CmpExists).
  if (cmp == PLCS_STR_CMP_PREFIX && val[0] == '\0') {
    return got ? PLCS_EVAL_RESULT_TRUE : PLCS_EVAL_RESULT_FALSE;
  }
  if (!got) {
    return PLCS_EVAL_RESULT_FALSE; // key absent
  }
  // Compare value-part (pattern) against the workload label value, matching
  // Go compareString(cmp, want, got): the default evaluator treats its first
  // argument as the pattern and second as the workload value.
  return plcs_default_string_evaluator(val, cmp, got, d, id);
}

// conf_reset clears the context and registers the host evaluators that make
// the C engine semantically identical to the Go engine for this corpus. Called
// once per vector.
static void conf_reset(void) {
  plcs_eval_ctx_reset();
  memset(conf_labels, 0, sizeof(conf_labels));
  memset(conf_str_params, 0, sizeof(conf_str_params));

  plcs_eval_ctx_register_str_evaluator(conf_always_true, PLCS_STR_EVAL_ALWAYS_TRUE);
  plcs_eval_ctx_register_str_evaluator(conf_always_false, PLCS_STR_EVAL_ALWAYS_FALSE);
  plcs_eval_ctx_register_str_evaluator(conf_always_abstain, PLCS_STR_EVAL_ALWAYS_ABSTAIN);

  plcs_eval_ctx_register_str_evaluator(conf_label, PLCS_STR_EVAL_NAMESPACE_LABEL);
  plcs_eval_ctx_register_str_evaluator(conf_label, PLCS_STR_EVAL_POD_LABEL);
  plcs_eval_ctx_register_str_evaluator(conf_label, PLCS_STR_EVAL_POD_ANNOTATION);
  plcs_eval_ctx_register_str_evaluator(conf_label, PLCS_STR_EVAL_CONTAINER_LABEL);

  for (int i = 1; i < PLCS_NUM_EVAL__COUNT; ++i) {
    plcs_eval_ctx_register_num_evaluator(conf_num, (plcs_numeric_evaluators)i);
    plcs_eval_ctx_register_unum_evaluator(conf_unum, (plcs_numeric_evaluators)i);
  }
}

static void conf_set_string(int id, const char *v) {
  if (id < 0 || id >= PLCS_STR_EVAL__COUNT) {
    return;
  }
  conf_copy(conf_str_params[id], v);
  plcs_eval_ctx_set_str_eval_param((plcs_string_evaluators)id, conf_str_params[id]);
}

static void conf_set_number(int id, long v) {
  plcs_eval_ctx_set_num_eval_param((plcs_numeric_evaluators)id, v);
}

static void conf_set_unumber(int id, unsigned long v) {
  plcs_eval_ctx_set_unum_eval_param((plcs_numeric_evaluators)id, v);
}

static void conf_label_present(int id) {
  if (id >= 0 && id < PLCS_STR_EVAL__COUNT) {
    conf_labels[id].present = 1;
  }
}

static void conf_add_label(int id, const char *k, const char *v) {
  if (id < 0 || id >= PLCS_STR_EVAL__COUNT) {
    return;
  }
  conf_label_set *s = &conf_labels[id];
  s->present = 1;
  if (s->count >= CONF_MAX_LABELS) {
    return;
  }
  conf_copy(s->keys[s->count], k);
  conf_copy(s->vals[s->count], v);
  s->count++;
}

static int conf_eval(const uint8_t *buf, size_t len) {
  (void)len;
  dd_wls_NodeTypeWrapper_table_t root = dd_wls_NodeTypeWrapper_as_root(buf);
  if (!root) {
    return (int)PLCS_EVAL_RESULT_ABSTAIN;
  }
  return (int)evaluate_rules(root, 0);
}
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/DataDog/dd-policy-engine/go/schema"
	"github.com/DataDog/dd-policy-engine/go/schema/dd/wls"
	flatbuffers "github.com/google/flatbuffers/go"
)

// dd-wls JSON identifiers, mirroring the wire enums decoded by
// policies/dd_wls.go. They are duplicated here (rather than imported) on
// purpose: this keeps the FlatBuffers/schema dependency confined to the
// conformance package and out of the importable "policies" package.
const (
	confNodeEvaluator = "EvaluatorNode"
	confNodeComposite = "CompositeNode"

	confEvalString   = "StrEvaluator"
	confEvalNumeric  = "NumEvaluator"
	confEvalUNumeric = "UNumEvaluator"
)

type confNodeWrap struct {
	NodeType string          `json:"node_type"`
	Node     json.RawMessage `json:"node"`
}

type confComposite struct {
	Description string         `json:"description"`
	Op          string         `json:"op"`
	Children    []confNodeWrap `json:"children"`
}

type confEvaluatorNode struct {
	Description string          `json:"description"`
	EvalType    string          `json:"eval_type"`
	Eval        json.RawMessage `json:"eval"`
}

type confStrEval struct {
	ID    string `json:"id"`
	Cmp   string `json:"cmp"`
	Value string `json:"value"`
}

type confNumEval struct {
	ID    string `json:"id"`
	Cmp   string `json:"cmp"`
	Value int64  `json:"value"`
}

type confUNumEval struct {
	ID    string `json:"id"`
	Cmp   string `json:"cmp"`
	Value uint64 `json:"value"`
}

// cResultName maps a C plcs_evaluation_result (TRUE=0, FALSE=1, ABSTAIN=2) to
// the corpus result name. Note the C enum order differs from the Go one, so
// the mapping must go through names rather than raw integers.
func cResultName(r C.int) string {
	switch r {
	case C.PLCS_EVAL_RESULT_TRUE:
		return "TRUE"
	case C.PLCS_EVAL_RESULT_FALSE:
		return "FALSE"
	default:
		return "ABSTAIN"
	}
}

// buildRulesBuffer serializes a vector's rule tree to a FlatBuffers buffer with
// a NodeTypeWrapper root, exactly as the C engine consumes it on the wire. It
// decodes the same dd-wls JSON the Go decoder reads, so both engines start from
// one source of truth.
func buildRulesBuffer(rules json.RawMessage) ([]byte, error) {
	var w confNodeWrap
	if err := json.Unmarshal(rules, &w); err != nil {
		return nil, err
	}
	b := flatbuffers.NewBuilder(256)
	off, err := buildNodeWrap(b, w)
	if err != nil {
		return nil, err
	}
	b.Finish(off)
	return b.FinishedBytes(), nil
}

func buildNodeWrap(b *flatbuffers.Builder, w confNodeWrap) (flatbuffers.UOffsetT, error) {
	switch w.NodeType {
	case confNodeComposite:
		var c confComposite
		if err := json.Unmarshal(w.Node, &c); err != nil {
			return 0, err
		}
		children := make([]flatbuffers.UOffsetT, 0, len(c.Children))
		for _, child := range c.Children {
			off, err := buildNodeWrap(b, child)
			if err != nil {
				return 0, err
			}
			children = append(children, off)
		}
		comp := schema.CompositeNodeCreate(b, wls.EnumValuesBoolOperation[c.Op], c.Description, children)
		return schema.NodeTypeWrapperCreate(b, comp, wls.NodeTypeCompositeNode), nil
	case confNodeEvaluator:
		var e confEvaluatorNode
		if err := json.Unmarshal(w.Node, &e); err != nil {
			return 0, err
		}
		var evalOff flatbuffers.UOffsetT
		var evalType wls.EvaluatorType
		switch e.EvalType {
		case confEvalString:
			var se confStrEval
			if err := json.Unmarshal(e.Eval, &se); err != nil {
				return 0, err
			}
			evalOff = schema.StrEvaluatorCreate(b, wls.EnumValuesStringEvaluators[se.ID], se.Value, wls.EnumValuesCmpTypeSTR[se.Cmp])
			evalType = wls.EvaluatorTypeStrEvaluator
		case confEvalNumeric:
			var ne confNumEval
			if err := json.Unmarshal(e.Eval, &ne); err != nil {
				return 0, err
			}
			evalOff = schema.NumEvaluatorCreate(b, wls.EnumValuesNumericEvaluators[ne.ID], ne.Value, wls.EnumValuesCmpTypeNUM[ne.Cmp])
			evalType = wls.EvaluatorTypeNumEvaluator
		case confEvalUNumeric:
			var ue confUNumEval
			if err := json.Unmarshal(e.Eval, &ue); err != nil {
				return 0, err
			}
			evalOff = schema.UNumEvaluatorCreate(b, wls.EnumValuesNumericEvaluators[ue.ID], ue.Value, wls.EnumValuesCmpTypeNUM[ue.Cmp])
			evalType = wls.EvaluatorTypeUNumEvaluator
		default:
			return 0, fmt.Errorf("unsupported eval_type %q", e.EvalType)
		}
		node := schema.EvaluatorNodeCreate(b, evalType, e.Description, evalOff)
		return schema.NodeTypeWrapperCreate(b, node, wls.NodeTypeEvaluatorNode), nil
	default:
		return 0, fmt.Errorf("unsupported node_type %q", w.NodeType)
	}
}

// cEvaluateBuffer loads the given facts into the C evaluation context and
// evaluates the serialized rule buffer with the real C engine, returning the
// tri-state result name (TRUE/FALSE/ABSTAIN).
func cEvaluateBuffer(buf []byte, strings map[string]string, labels map[string]map[string]string, numbers map[string]int64, unumbers map[string]uint64) string {
	C.conf_reset()

	for id, val := range strings {
		cs := C.CString(val)
		C.conf_set_string(C.int(wls.EnumValuesStringEvaluators[id]), cs)
		C.free(unsafe.Pointer(cs))
	}
	for id, m := range labels {
		cid := C.int(wls.EnumValuesStringEvaluators[id])
		C.conf_label_present(cid)
		for k, val := range m {
			ck := C.CString(k)
			cv := C.CString(val)
			C.conf_add_label(cid, ck, cv)
			C.free(unsafe.Pointer(ck))
			C.free(unsafe.Pointer(cv))
		}
	}
	for id, val := range numbers {
		C.conf_set_number(C.int(wls.EnumValuesNumericEvaluators[id]), C.long(val))
	}
	for id, val := range unumbers {
		C.conf_set_unumber(C.int(wls.EnumValuesNumericEvaluators[id]), C.ulong(val))
	}

	var p *C.uint8_t
	if len(buf) > 0 {
		p = (*C.uint8_t)(unsafe.Pointer(&buf[0]))
	}
	return cResultName(C.conf_eval(p, C.size_t(len(buf))))
}

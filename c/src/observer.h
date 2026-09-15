/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache 2.0 License. This product includes software developed at
 * Datadog (https://www.datadoghq.com/).
 *
 * Copyright 2025-Present Datadog, Inc.
 */
#pragma once

#include <policy_reader.h>

#include <dd/policies/evaluation_result.h>
#include <dd/policies/observer.h>

#include "wire/dd_types.h"

/**
 * @brief Describes a node an evaluation just walked, for an observer to report on.
 *
 * @param node    the node, either an evaluator or a composite.
 * @param result  what it evaluated to.
 * @param depth   how deep it sits in the policy's rule tree.
 */
plcs_evaluation_record
plcs_describe_node(dd_ns(NodeTypeWrapper_table_t) node, plcs_evaluation_result result, int depth);

// Copyright (c) 2026 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Apache Software License 2.0 which is available at
// https://www.apache.org/licenses/LICENSE-2.0, or the MIT license
// which is available at https://opensource.org/licenses/MIT.
//
// SPDX-License-Identifier: Apache-2.0 OR MIT

package iox2

// #include "iox2/iceoryx2.h"
import "C"

import (
	"errors"
	"fmt"
)

// ErrInvalidServiceType is returned when a [ServiceType] has a value that is
// not defined by this package.
var ErrInvalidServiceType = errors.New("iox2: invalid service type")

// SemanticStringError describes why a string, for example a node name, was
// rejected by iceoryx2.
type SemanticStringError int

// The reasons for which iceoryx2 rejects a string.
const (
	// SemanticStringErrorInvalidContent means the string contains characters
	// that are not allowed.
	SemanticStringErrorInvalidContent SemanticStringError = C.iox2_semantic_string_error_e_INVALID_CONTENT
	// SemanticStringErrorExceedsMaximumLength means the string is longer than
	// the maximum length.
	SemanticStringErrorExceedsMaximumLength SemanticStringError = C.iox2_semantic_string_error_e_EXCEEDS_MAXIMUM_LENGTH
)

// Error implements the error interface.
func (e SemanticStringError) Error() string {
	switch e {
	case SemanticStringErrorInvalidContent, SemanticStringErrorExceedsMaximumLength:
		return "iox2: " + C.GoString(C.iox2_semantic_string_error_string(uint32(e)))
	default:
		return fmt.Sprintf("iox2: unknown semantic string error %d", int(e))
	}
}

// NodeCreationFailure describes why iceoryx2 was unable to create a [Node].
type NodeCreationFailure int

// The reasons for which creating a [Node] can fail.
const (
	// NodeCreationFailureInsufficientPermissions means the process lacks the
	// permissions to create the node resources.
	NodeCreationFailureInsufficientPermissions NodeCreationFailure = C.iox2_node_creation_failure_e_INSUFFICIENT_PERMISSIONS
	// NodeCreationFailureInternalError means iceoryx2 hit an internal error.
	NodeCreationFailureInternalError NodeCreationFailure = C.iox2_node_creation_failure_e_INTERNAL_ERROR
	// NodeCreationFailureSystemCorrupted means the iceoryx2 system resources
	// are in an inconsistent state.
	NodeCreationFailureSystemCorrupted NodeCreationFailure = C.iox2_node_creation_failure_e_SYSTEM_CORRUPTED
	// NodeCreationFailureUnableToGenerateUniqueNodeID means no unique node id
	// could be generated.
	NodeCreationFailureUnableToGenerateUniqueNodeID NodeCreationFailure = C.iox2_node_creation_failure_e_UNABLE_TO_GENERATE_UNIQUE_NODE_ID
)

// Error implements the error interface.
func (e NodeCreationFailure) Error() string {
	switch e {
	case NodeCreationFailureInsufficientPermissions,
		NodeCreationFailureInternalError,
		NodeCreationFailureSystemCorrupted,
		NodeCreationFailureUnableToGenerateUniqueNodeID:
		return "iox2: " + C.GoString(C.iox2_node_creation_failure_string(uint32(e)))
	default:
		return fmt.Sprintf("iox2: unknown node creation failure %d", int(e))
	}
}

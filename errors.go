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

// ErrClosed is returned when a node, service, publisher, subscriber or sample
// is used after it was closed.
var ErrClosed = errors.New("iox2: use of closed object")

// PubSubServiceError describes why a publish-subscribe service could not be
// opened or created. The values are the ones a caller commonly distinguishes,
// every other reason is reported with the message of the C API.
type PubSubServiceError int

// The reasons for which opening or creating a publish-subscribe service fails
// that callers commonly react to.
const (
	// PubSubServiceErrorDoesNotExist means [PubSubServiceBuilder.Open] found no
	// service of that name.
	PubSubServiceErrorDoesNotExist PubSubServiceError = C.iox2_pub_sub_open_or_create_error_e_O_DOES_NOT_EXIST
	// PubSubServiceErrorIncompatibleTypes means the service exists with a
	// different payload type.
	PubSubServiceErrorIncompatibleTypes PubSubServiceError = C.iox2_pub_sub_open_or_create_error_e_O_INCOMPATIBLE_TYPES
	// PubSubServiceErrorIncompatibleMessagingPattern means a service of that
	// name exists, but it is not a publish-subscribe service.
	PubSubServiceErrorIncompatibleMessagingPattern PubSubServiceError = C.iox2_pub_sub_open_or_create_error_e_O_INCOMPATIBLE_MESSAGING_PATTERN
	// PubSubServiceErrorInsufficientPermissions means the process may not
	// open the service.
	PubSubServiceErrorInsufficientPermissions PubSubServiceError = C.iox2_pub_sub_open_or_create_error_e_O_INSUFFICIENT_PERMISSIONS
	// PubSubServiceErrorAlreadyExists means [PubSubServiceBuilder.Create] found
	// a service of that name.
	PubSubServiceErrorAlreadyExists PubSubServiceError = C.iox2_pub_sub_open_or_create_error_e_C_ALREADY_EXISTS
)

// Error implements the error interface.
func (e PubSubServiceError) Error() string {
	// The string function is only defined for the values of the enum, which are
	// consecutive up to the last one.
	if e >= C.iox2_pub_sub_open_or_create_error_e_O_DOES_NOT_EXIST &&
		e <= C.iox2_pub_sub_open_or_create_error_e_C_UNABLE_TO_GENERATE_UNIQUE_SERVICE_ID {
		return "iox2: " + C.GoString(C.iox2_pub_sub_open_or_create_error_string(uint32(e)))
	}
	return fmt.Sprintf("iox2: unknown publish-subscribe service error %d", int(e))
}

// PublisherCreateError describes why a [Publisher] could not be created.
type PublisherCreateError int

// The reasons for which creating a [Publisher] can fail.
const (
	// PublisherCreateErrorExceedsMaxSupportedPublishers means the service
	// already has as many publishers as it supports.
	PublisherCreateErrorExceedsMaxSupportedPublishers PublisherCreateError = C.iox2_publisher_create_error_e_EXCEEDS_MAX_SUPPORTED_PUBLISHERS
	// PublisherCreateErrorUnableToCreateDataSegment means the shared memory for
	// the samples could not be created.
	PublisherCreateErrorUnableToCreateDataSegment PublisherCreateError = C.iox2_publisher_create_error_e_UNABLE_TO_CREATE_DATA_SEGMENT
	// PublisherCreateErrorFailedToDeployThreadSafetyPolicy means the thread
	// safety policy of the port could not be set up.
	PublisherCreateErrorFailedToDeployThreadSafetyPolicy PublisherCreateError = C.iox2_publisher_create_error_e_FAILED_TO_DEPLOY_THREAD_SAFETY_POLICY
	// PublisherCreateErrorUnableToCreatePortTag means the tag that marks the
	// port as alive could not be created.
	PublisherCreateErrorUnableToCreatePortTag PublisherCreateError = C.iox2_publisher_create_error_e_UNABLE_TO_CREATE_PORT_TAG
	// PublisherCreateErrorUnableToGenerateUniquePublisherID means no unique
	// publisher id could be generated.
	PublisherCreateErrorUnableToGenerateUniquePublisherID PublisherCreateError = C.iox2_publisher_create_error_e_UNABLE_TO_GENERATE_UNIQUE_PUBLISHER_ID
)

// Error implements the error interface.
func (e PublisherCreateError) Error() string {
	if e >= C.iox2_publisher_create_error_e_EXCEEDS_MAX_SUPPORTED_PUBLISHERS &&
		e <= C.iox2_publisher_create_error_e_UNABLE_TO_GENERATE_UNIQUE_PUBLISHER_ID {
		return "iox2: " + C.GoString(C.iox2_publisher_create_error_string(uint32(e)))
	}
	return fmt.Sprintf("iox2: unknown publisher creation error %d", int(e))
}

// SubscriberCreateError describes why a [Subscriber] could not be created.
type SubscriberCreateError int

// The reasons for which creating a [Subscriber] can fail.
const (
	// SubscriberCreateErrorExceedsMaxSupportedSubscribers means the service
	// already has as many subscribers as it supports.
	SubscriberCreateErrorExceedsMaxSupportedSubscribers SubscriberCreateError = C.iox2_subscriber_create_error_e_EXCEEDS_MAX_SUPPORTED_SUBSCRIBERS
	// SubscriberCreateErrorBufferSizeExceedsMaxSupportedBufferSizeOfService
	// means the requested buffer is larger than the service supports.
	SubscriberCreateErrorBufferSizeExceedsMaxSupportedBufferSizeOfService SubscriberCreateError = C.iox2_subscriber_create_error_e_BUFFER_SIZE_EXCEEDS_MAX_SUPPORTED_BUFFER_SIZE_OF_SERVICE
	// SubscriberCreateErrorFailedToDeployThreadSafetyPolicy means the thread
	// safety policy of the port could not be set up.
	SubscriberCreateErrorFailedToDeployThreadSafetyPolicy SubscriberCreateError = C.iox2_subscriber_create_error_e_FAILED_TO_DEPLOY_THREAD_SAFETY_POLICY
	// SubscriberCreateErrorUnableToCreatePortTag means the tag that marks the
	// port as alive could not be created.
	SubscriberCreateErrorUnableToCreatePortTag SubscriberCreateError = C.iox2_subscriber_create_error_e_UNABLE_TO_CREATE_PORT_TAG
	// SubscriberCreateErrorHistoryRequestExceedsHistorySizeOfService means the
	// requested history is larger than the history of the service.
	SubscriberCreateErrorHistoryRequestExceedsHistorySizeOfService SubscriberCreateError = C.iox2_subscriber_create_error_e_HISTORY_REQUEST_EXCEEDS_HISTORY_SIZE_OF_SERVICE
	// SubscriberCreateErrorHistoryRequestExceedsBufferSizeOfSubscriber means
	// the requested history is larger than the buffer of the subscriber.
	SubscriberCreateErrorHistoryRequestExceedsBufferSizeOfSubscriber SubscriberCreateError = C.iox2_subscriber_create_error_e_HISTORY_REQUEST_EXCEEDS_BUFFER_SIZE_OF_SUBSCRIBER
	// SubscriberCreateErrorUnableToGenerateUniqueSubscriberID means no unique
	// subscriber id could be generated.
	SubscriberCreateErrorUnableToGenerateUniqueSubscriberID SubscriberCreateError = C.iox2_subscriber_create_error_e_UNABLE_TO_GENERATE_UNIQUE_SUBSCRIBER_ID
)

// Error implements the error interface.
func (e SubscriberCreateError) Error() string {
	if e >= C.iox2_subscriber_create_error_e_EXCEEDS_MAX_SUPPORTED_SUBSCRIBERS &&
		e <= C.iox2_subscriber_create_error_e_UNABLE_TO_GENERATE_UNIQUE_SUBSCRIBER_ID {
		return "iox2: " + C.GoString(C.iox2_subscriber_create_error_string(uint32(e)))
	}
	return fmt.Sprintf("iox2: unknown subscriber creation error %d", int(e))
}

// LoanError describes why a [Publisher] could not obtain memory for a sample.
type LoanError int

// The reasons for which loaning memory for a sample can fail.
const (
	// LoanErrorOutOfMemory means the data segment of the publisher is exhausted.
	LoanErrorOutOfMemory LoanError = C.iox2_loan_error_e_OUT_OF_MEMORY
	// LoanErrorExceedsMaxLoanedSamples means the publisher already holds as
	// many loaned samples as it may.
	LoanErrorExceedsMaxLoanedSamples LoanError = C.iox2_loan_error_e_EXCEEDS_MAX_LOANED_SAMPLES
	// LoanErrorExceedsMaxLoanSize means the payload is longer than the data
	// segment of the publisher can hold.
	LoanErrorExceedsMaxLoanSize LoanError = C.iox2_loan_error_e_EXCEEDS_MAX_LOAN_SIZE
	// LoanErrorInternalFailure means iceoryx2 hit an internal error.
	LoanErrorInternalFailure LoanError = C.iox2_loan_error_e_INTERNAL_FAILURE
)

// Error implements the error interface.
func (e LoanError) Error() string {
	if e >= C.iox2_loan_error_e_OUT_OF_MEMORY && e <= C.iox2_loan_error_e_INTERNAL_FAILURE {
		return "iox2: " + C.GoString(C.iox2_loan_error_string(uint32(e)))
	}
	return fmt.Sprintf("iox2: unknown loan error %d", int(e))
}

// SendError describes why a sample could not be delivered to the subscribers.
type SendError int

// The reasons for which sending a sample can fail.
const (
	// SendErrorConnectionBrokenSinceSenderNoLongerExists means the connection
	// to a subscriber broke because the publisher is gone.
	SendErrorConnectionBrokenSinceSenderNoLongerExists SendError = C.iox2_send_error_e_CONNECTION_BROKEN_SINCE_SENDER_NO_LONGER_EXISTS
	// SendErrorConnectionCorrupted means the connection to a subscriber is in
	// an inconsistent state.
	SendErrorConnectionCorrupted SendError = C.iox2_send_error_e_CONNECTION_CORRUPTED
	// SendErrorLoanErrorOutOfMemory means the data segment is exhausted.
	SendErrorLoanErrorOutOfMemory SendError = C.iox2_send_error_e_LOAN_ERROR_OUT_OF_MEMORY
	// SendErrorLoanErrorExceedsMaxLoans means too many samples are loaned.
	SendErrorLoanErrorExceedsMaxLoans SendError = C.iox2_send_error_e_LOAN_ERROR_EXCEEDS_MAX_LOANS
	// SendErrorLoanErrorExceedsMaxLoanSize means the payload is too long.
	SendErrorLoanErrorExceedsMaxLoanSize SendError = C.iox2_send_error_e_LOAN_ERROR_EXCEEDS_MAX_LOAN_SIZE
	// SendErrorLoanErrorInternalFailure means iceoryx2 hit an internal error
	// while loaning memory.
	SendErrorLoanErrorInternalFailure SendError = C.iox2_send_error_e_LOAN_ERROR_INTERNAL_FAILURE
	// SendErrorConnectionError means a connection to a subscriber could not
	// be established.
	SendErrorConnectionError SendError = C.iox2_send_error_e_CONNECTION_ERROR
	// SendErrorUnableToDeliver means the sample could not be delivered.
	SendErrorUnableToDeliver SendError = C.iox2_send_error_e_UNABLE_TO_DELIVER
	// SendErrorInternalError means iceoryx2 hit an internal error.
	SendErrorInternalError SendError = C.iox2_send_error_e_INTERNAL_ERROR
)

// Error implements the error interface.
func (e SendError) Error() string {
	if e >= C.iox2_send_error_e_CONNECTION_BROKEN_SINCE_SENDER_NO_LONGER_EXISTS &&
		e <= C.iox2_send_error_e_INTERNAL_ERROR {
		return "iox2: " + C.GoString(C.iox2_send_error_string(uint32(e)))
	}
	return fmt.Sprintf("iox2: unknown send error %d", int(e))
}

// ReceiveError describes why a [Subscriber] could not take a sample from its
// queue. An empty queue is not an error.
type ReceiveError int

// The reasons for which receiving a sample can fail.
const (
	// ReceiveErrorExceedsMaxBorrows means the subscriber holds as many samples
	// as it may. Close some of them.
	ReceiveErrorExceedsMaxBorrows ReceiveError = C.iox2_receive_error_e_EXCEEDS_MAX_BORROWS
	// ReceiveErrorFailedToEstablishConnection means the connection to a
	// publisher could not be established.
	ReceiveErrorFailedToEstablishConnection ReceiveError = C.iox2_receive_error_e_FAILED_TO_ESTABLISH_CONNECTION
	// ReceiveErrorUnableToMapSendersDataSegment means the shared memory of a
	// publisher could not be mapped.
	ReceiveErrorUnableToMapSendersDataSegment ReceiveError = C.iox2_receive_error_e_UNABLE_TO_MAP_SENDERS_DATA_SEGMENT
)

// Error implements the error interface.
func (e ReceiveError) Error() string {
	if e >= C.iox2_receive_error_e_EXCEEDS_MAX_BORROWS &&
		e <= C.iox2_receive_error_e_UNABLE_TO_MAP_SENDERS_DATA_SEGMENT {
		return "iox2: " + C.GoString(C.iox2_receive_error_string(uint32(e)))
	}
	return fmt.Sprintf("iox2: unknown receive error %d", int(e))
}

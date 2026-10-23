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

package iox2_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	iox2 "github.com/eclipse-iceoryx/iceoryx2-go"
)

// reading is a fixed-size payload. Its layout is that of a C struct with the
// same fields.
type reading struct {
	ID    uint32
	Flags uint8
	Value float64
}

const readingTypeName = "iox2_go_test::Reading"

// serviceName returns a service name that is unique to the test and the
// process, so that tests do not interfere with each other or with services that
// a crashed run left behind.
func serviceName(t *testing.T) string {
	t.Helper()
	name := strings.NewReplacer(" ", "_", "#", "_").Replace(t.Name())
	return fmt.Sprintf("iox2-go-test/%s/%d", name, os.Getpid())
}

func newNode(t *testing.T, serviceType iox2.ServiceType) *iox2.Node {
	t.Helper()
	node, err := iox2.NewNodeBuilder().Name("go-test").Create(serviceType)
	if err != nil {
		t.Fatalf("Create node: %v", err)
	}
	t.Cleanup(node.Close)
	return node
}

// ports opens the service of the test and creates one publisher and one
// subscriber on it.
func ports(t *testing.T, node *iox2.Node, payload iox2.PayloadType) (*iox2.Publisher, *iox2.Subscriber) {
	t.Helper()
	service, err := node.PublishSubscribe(serviceName(t)).Payload(payload).OpenOrCreate()
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	t.Cleanup(service.Close)
	publisher, err := service.PublisherBuilder().Create()
	if err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	t.Cleanup(publisher.Close)
	subscriber, err := service.SubscriberBuilder().Create()
	if err != nil {
		t.Fatalf("create subscriber: %v", err)
	}
	t.Cleanup(subscriber.Close)
	return publisher, subscriber
}

func receiveBytes(t *testing.T, subscriber *iox2.Subscriber) []byte {
	t.Helper()
	sample, ok, err := subscriber.Receive()
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if !ok {
		t.Fatal("Receive found no sample")
	}
	defer sample.Close()
	return bytes.Clone(sample.Payload())
}

func TestBytesRoundTrip(t *testing.T) {
	for _, serviceType := range serviceTypes {
		t.Run(serviceType.String(), func(t *testing.T) {
			publisher, subscriber := ports(t, newNode(t, serviceType), iox2.BytesPayload())

			if _, ok, err := subscriber.Receive(); ok || err != nil {
				t.Fatalf("Receive on an empty queue = %v, %v, want no sample and no error", ok, err)
			}

			// The sizes cross the boundaries at which the data segment grows.
			for _, size := range []int{1, 7, 64, 1000, 4096, 100000} {
				want := bytes.Repeat([]byte{byte(size)}, size)
				want[size-1] = 0xAB
				recipients, err := publisher.Send(want)
				if err != nil {
					t.Fatalf("Send %d bytes: %v", size, err)
				}
				if recipients != 1 {
					t.Errorf("Send %d bytes: %d recipients, want 1", size, recipients)
				}
				if got := receiveBytes(t, subscriber); !bytes.Equal(got, want) {
					t.Errorf("received %d bytes that differ from the %d bytes sent", len(got), size)
				}
			}
		})
	}
}

func TestSamplesArriveInOrder(t *testing.T) {
	publisher, subscriber := ports(t, newNode(t, iox2.ServiceTypeLocal), iox2.BytesPayload())

	// The default buffer of a subscriber holds two samples.
	for _, msg := range []string{"first", "second"} {
		if _, err := publisher.Send([]byte(msg)); err != nil {
			t.Fatalf("Send: %v", err)
		}
	}
	for _, want := range []string{"first", "second"} {
		if got := string(receiveBytes(t, subscriber)); got != want {
			t.Errorf("received %q, want %q", got, want)
		}
	}
}

func TestSendWithoutSubscriberHasNoRecipients(t *testing.T) {
	service, err := newNode(t, iox2.ServiceTypeLocal).PublishSubscribe(serviceName(t)).OpenOrCreate()
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	defer service.Close()
	publisher, err := service.PublisherBuilder().Create()
	if err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	defer publisher.Close()

	recipients, err := publisher.Send([]byte("nobody listens"))
	if err != nil || recipients != 0 {
		t.Errorf("Send = %d, %v, want 0 recipients and no error", recipients, err)
	}
}

func TestEmptyBytesPayload(t *testing.T) {
	publisher, subscriber := ports(t, newNode(t, iox2.ServiceTypeLocal), iox2.BytesPayload())

	if _, err := publisher.Send(nil); err != nil {
		t.Fatalf("Send(nil): %v", err)
	}
	if got := receiveBytes(t, subscriber); len(got) != 0 {
		t.Errorf("received %d bytes, want none", len(got))
	}
}

func TestFixedSizeRoundTrip(t *testing.T) {
	for _, serviceType := range serviceTypes {
		t.Run(serviceType.String(), func(t *testing.T) {
			publisher, subscriber := ports(t, newNode(t, serviceType), iox2.FixedSizePayload[reading](readingTypeName))

			want := reading{ID: 42, Flags: 7, Value: 3.25}
			if recipients, err := iox2.SendValue(publisher, want); err != nil || recipients != 1 {
				t.Fatalf("SendValue = %d, %v, want 1 recipient and no error", recipients, err)
			}
			got, ok, err := iox2.ReceiveValue[reading](subscriber)
			if err != nil || !ok {
				t.Fatalf("ReceiveValue = %v, %v, want a value", ok, err)
			}
			if got != want {
				t.Errorf("ReceiveValue = %+v, want %+v", got, want)
			}

			if _, ok, err := iox2.ReceiveValue[reading](subscriber); ok || err != nil {
				t.Errorf("ReceiveValue on an empty queue = %v, %v, want no value and no error", ok, err)
			}
		})
	}
}

func TestFixedSizeAsBytes(t *testing.T) {
	publisher, subscriber := ports(t, newNode(t, iox2.ServiceTypeLocal), iox2.FixedSizePayload[[4]uint16]("[u16; 4]"))

	want := []byte{1, 0, 2, 0, 3, 0, 4, 0}
	if _, err := publisher.Send(want); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got := receiveBytes(t, subscriber); !bytes.Equal(got, want) {
		t.Errorf("received %v, want %v", got, want)
	}

	for _, size := range []int{0, 7, 9} {
		if _, err := publisher.Send(make([]byte, size)); !errors.Is(err, iox2.ErrPayloadTypeMismatch) {
			t.Errorf("Send(%d bytes) error = %v, want %v", size, err, iox2.ErrPayloadTypeMismatch)
		}
	}
}

func TestTypedCallsCheckTheType(t *testing.T) {
	type other struct{ A, B, C uint64 } // same size as reading
	publisher, subscriber := ports(t, newNode(t, iox2.ServiceTypeLocal), iox2.FixedSizePayload[reading](readingTypeName))

	if _, err := iox2.SendValue(publisher, other{}); !errors.Is(err, iox2.ErrPayloadTypeMismatch) {
		t.Errorf("SendValue of another type: error = %v, want %v", err, iox2.ErrPayloadTypeMismatch)
	}
	if _, err := iox2.SendValue(publisher, uint8(1)); !errors.Is(err, iox2.ErrPayloadTypeMismatch) {
		t.Errorf("SendValue of a smaller type: error = %v, want %v", err, iox2.ErrPayloadTypeMismatch)
	}
	if _, _, err := iox2.ReceiveValue[other](subscriber); !errors.Is(err, iox2.ErrPayloadTypeMismatch) {
		t.Errorf("ReceiveValue of another type: error = %v, want %v", err, iox2.ErrPayloadTypeMismatch)
	}
}

func TestBytesServiceRejectsTypedCalls(t *testing.T) {
	publisher, subscriber := ports(t, newNode(t, iox2.ServiceTypeLocal), iox2.BytesPayload())

	if _, err := iox2.SendValue(publisher, uint8(1)); !errors.Is(err, iox2.ErrPayloadTypeMismatch) {
		t.Errorf("SendValue: error = %v, want %v", err, iox2.ErrPayloadTypeMismatch)
	}
	if _, _, err := iox2.ReceiveValue[uint8](subscriber); !errors.Is(err, iox2.ErrPayloadTypeMismatch) {
		t.Errorf("ReceiveValue: error = %v, want %v", err, iox2.ErrPayloadTypeMismatch)
	}
}

func TestInvalidPayloadTypesAreRejected(t *testing.T) {
	type withPointer struct{ P *int }
	type withString struct{ S string }
	type withSlice struct{ S []byte }
	type withInt struct{ N int }
	type nested struct {
		A [2]struct{ M map[string]int }
	}
	tests := []struct {
		name    string
		payload iox2.PayloadType
	}{
		{"pointer", iox2.FixedSizePayload[withPointer]("t")},
		{"string", iox2.FixedSizePayload[withString]("t")},
		{"slice", iox2.FixedSizePayload[withSlice]("t")},
		{"int", iox2.FixedSizePayload[withInt]("t")},
		{"nested map", iox2.FixedSizePayload[nested]("t")},
		{"interface", iox2.FixedSizePayload[any]("t")},
		{"zero size", iox2.FixedSizePayload[struct{}]("t")},
		{"empty type name", iox2.FixedSizePayload[reading]("")},
		{"type name too long", iox2.FixedSizePayload[reading](strings.Repeat("n", 257))},
		{"type name not utf-8", iox2.FixedSizePayload[reading]("\xff\xfe")},
		{"zero value", iox2.PayloadType{}},
	}
	node := newNode(t, iox2.ServiceTypeLocal)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := node.PublishSubscribe(serviceName(t)).Payload(tt.payload).OpenOrCreate()
			if service != nil {
				service.Close()
				t.Fatal("OpenOrCreate returned a service")
			}
			if !errors.Is(err, iox2.ErrInvalidPayloadType) {
				t.Errorf("OpenOrCreate error = %v, want %v", err, iox2.ErrInvalidPayloadType)
			}
		})
	}
}

func TestPayloadTypeAccessors(t *testing.T) {
	p := iox2.FixedSizePayload[reading](readingTypeName)
	if p.TypeName() != readingTypeName || p.Size() != 16 || p.Alignment() != 8 {
		t.Errorf("payload = %q, size %d, alignment %d, want %q, 16, 8", p.TypeName(), p.Size(), p.Alignment(), readingTypeName)
	}
	b := iox2.BytesPayload()
	if b.TypeName() != "u8" || b.Size() != 1 || b.Alignment() != 1 {
		t.Errorf("bytes payload = %q, size %d, alignment %d, want \"u8\", 1, 1", b.TypeName(), b.Size(), b.Alignment())
	}
}

func TestServiceOpenAndCreate(t *testing.T) {
	for _, serviceType := range serviceTypes {
		t.Run(serviceType.String(), func(t *testing.T) {
			node := newNode(t, serviceType)
			name := serviceName(t)

			if _, err := node.PublishSubscribe(name).Open(); !errors.Is(err, iox2.PubSubServiceErrorDoesNotExist) {
				t.Fatalf("Open of a missing service: error = %v, want %v", err, iox2.PubSubServiceErrorDoesNotExist)
			}
			created, err := node.PublishSubscribe(name).Create()
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			defer created.Close()
			if created.Name() != name || created.Payload().TypeName() != "u8" {
				t.Errorf("service = %q with payload %q, want %q with \"u8\"", created.Name(), created.Payload().TypeName(), name)
			}

			if _, err := node.PublishSubscribe(name).Create(); !errors.Is(err, iox2.PubSubServiceErrorAlreadyExists) {
				t.Errorf("second Create: error = %v, want %v", err, iox2.PubSubServiceErrorAlreadyExists)
			}
			opened, err := node.PublishSubscribe(name).Open()
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			opened.Close()
		})
	}
}

func TestOpenWithOtherPayloadTypeFails(t *testing.T) {
	node := newNode(t, iox2.ServiceTypeLocal)
	name := serviceName(t)
	created, err := node.PublishSubscribe(name).Payload(iox2.FixedSizePayload[reading](readingTypeName)).Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer created.Close()

	others := map[string]iox2.PayloadType{
		"bytes":     iox2.BytesPayload(),
		"type name": iox2.FixedSizePayload[reading]("another::Reading"),
		"size":      iox2.FixedSizePayload[[2]reading](readingTypeName),
	}
	for what, payload := range others {
		service, err := node.PublishSubscribe(name).Payload(payload).Open()
		if service != nil {
			service.Close()
		}
		if !errors.Is(err, iox2.PubSubServiceErrorIncompatibleTypes) {
			t.Errorf("Open with another %s: error = %v, want %v", what, err, iox2.PubSubServiceErrorIncompatibleTypes)
		}
	}
}

func TestInvalidServiceNames(t *testing.T) {
	node := newNode(t, iox2.ServiceTypeLocal)
	tests := map[string]struct {
		name string
		want iox2.SemanticStringError
	}{
		"empty":        {"", iox2.SemanticStringErrorInvalidContent},
		"too long":     {strings.Repeat("s", 256), iox2.SemanticStringErrorExceedsMaximumLength},
		"embedded nul": {"go\x00service", iox2.SemanticStringErrorInvalidContent},
	}
	for what, tt := range tests {
		t.Run(what, func(t *testing.T) {
			service, err := node.PublishSubscribe(tt.name).OpenOrCreate()
			if service != nil {
				service.Close()
				t.Fatal("OpenOrCreate returned a service")
			}
			var got iox2.SemanticStringError
			if !errors.As(err, &got) || got != tt.want {
				t.Errorf("OpenOrCreate error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestPortLimitsAreReported(t *testing.T) {
	service, err := newNode(t, iox2.ServiceTypeLocal).PublishSubscribe(serviceName(t)).OpenOrCreate()
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	defer service.Close()

	// By default a service has two publishers and eight subscribers at most.
	var publishers []*iox2.Publisher
	defer func() {
		for _, p := range publishers {
			p.Close()
		}
	}()
	var err2 error
	for range 100 {
		var p *iox2.Publisher
		if p, err2 = service.PublisherBuilder().Create(); err2 != nil {
			break
		}
		publishers = append(publishers, p)
	}
	if !errors.Is(err2, iox2.PublisherCreateErrorExceedsMaxSupportedPublishers) {
		t.Errorf("creating publishers: error = %v, want %v", err2, iox2.PublisherCreateErrorExceedsMaxSupportedPublishers)
	}

	var subscribers []*iox2.Subscriber
	defer func() {
		for _, s := range subscribers {
			s.Close()
		}
	}()
	for range 100 {
		var s *iox2.Subscriber
		if s, err2 = service.SubscriberBuilder().Create(); err2 != nil {
			break
		}
		subscribers = append(subscribers, s)
	}
	if !errors.Is(err2, iox2.SubscriberCreateErrorExceedsMaxSupportedSubscribers) {
		t.Errorf("creating subscribers: error = %v, want %v", err2, iox2.SubscriberCreateErrorExceedsMaxSupportedSubscribers)
	}
}

func TestSubscriberLimitsBorrowedSamples(t *testing.T) {
	publisher, subscriber := ports(t, newNode(t, iox2.ServiceTypeLocal), iox2.BytesPayload())

	var held []*iox2.Sample
	defer func() {
		for _, s := range held {
			s.Close()
		}
	}()
	var err error
	for range 100 {
		if _, err = publisher.Send([]byte("x")); err != nil {
			t.Fatalf("Send: %v", err)
		}
		var sample *iox2.Sample
		var ok bool
		if sample, ok, err = subscriber.Receive(); err != nil || !ok {
			break
		}
		held = append(held, sample)
	}
	if !errors.Is(err, iox2.ReceiveErrorExceedsMaxBorrows) {
		t.Fatalf("Receive error after %d held samples = %v, want %v", len(held), err, iox2.ReceiveErrorExceedsMaxBorrows)
	}

	// Releasing samples makes room again.
	held[0].Close()
	held = held[1:]
	sample, ok, err := subscriber.Receive()
	if err != nil || !ok {
		t.Fatalf("Receive after Close = %v, %v, want a sample", ok, err)
	}
	sample.Close()
}

func TestErrorMessagesComeFromTheCAPI(t *testing.T) {
	errs := []error{
		iox2.PubSubServiceErrorDoesNotExist,
		iox2.PubSubServiceErrorAlreadyExists,
		iox2.PublisherCreateErrorExceedsMaxSupportedPublishers,
		iox2.SubscriberCreateErrorExceedsMaxSupportedSubscribers,
		iox2.LoanErrorExceedsMaxLoanSize,
		iox2.SendErrorUnableToDeliver,
		iox2.ReceiveErrorExceedsMaxBorrows,
	}
	for _, err := range errs {
		if msg := err.Error(); !strings.HasPrefix(msg, "iox2: ") || strings.Contains(msg, "unknown") {
			t.Errorf("Error() = %q, want a message from the C API", msg)
		}
	}
	unknown := []error{
		iox2.PubSubServiceError(-1), iox2.PubSubServiceError(1000),
		iox2.PublisherCreateError(1000), iox2.SubscriberCreateError(1000),
		iox2.LoanError(1000), iox2.SendError(1000), iox2.ReceiveError(1000),
	}
	for _, err := range unknown {
		if !strings.Contains(err.Error(), "unknown") {
			t.Errorf("Error() = %q, want it to report an unknown value", err.Error())
		}
	}
}

func TestClosedObjectsReportErrClosed(t *testing.T) {
	node := newNode(t, iox2.ServiceTypeLocal)
	service, err := node.PublishSubscribe(serviceName(t)).OpenOrCreate()
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	publisher, err := service.PublisherBuilder().Create()
	if err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	subscriber, err := service.SubscriberBuilder().Create()
	if err != nil {
		t.Fatalf("create subscriber: %v", err)
	}
	if _, err := publisher.Send([]byte("data")); err != nil {
		t.Fatalf("Send: %v", err)
	}
	sample, ok, err := subscriber.Receive()
	if err != nil || !ok {
		t.Fatalf("Receive = %v, %v, want a sample", ok, err)
	}

	for range 2 { // Close is idempotent
		sample.Close()
		publisher.Close()
		subscriber.Close()
		service.Close()
	}

	if payload := sample.Payload(); payload != nil {
		t.Errorf("Payload of a closed sample = %v, want nil", payload)
	}
	if _, err := publisher.Send([]byte("data")); !errors.Is(err, iox2.ErrClosed) {
		t.Errorf("Send after Close: error = %v, want %v", err, iox2.ErrClosed)
	}
	if _, _, err := subscriber.Receive(); !errors.Is(err, iox2.ErrClosed) {
		t.Errorf("Receive after Close: error = %v, want %v", err, iox2.ErrClosed)
	}
	if _, err := service.PublisherBuilder().Create(); !errors.Is(err, iox2.ErrClosed) {
		t.Errorf("creating a publisher after Close: error = %v, want %v", err, iox2.ErrClosed)
	}
	if _, err := service.SubscriberBuilder().Create(); !errors.Is(err, iox2.ErrClosed) {
		t.Errorf("creating a subscriber after Close: error = %v, want %v", err, iox2.ErrClosed)
	}

	node.Close()
	if _, err := node.PublishSubscribe(serviceName(t)).OpenOrCreate(); !errors.Is(err, iox2.ErrClosed) {
		t.Errorf("creating a service on a closed node: error = %v, want %v", err, iox2.ErrClosed)
	}
}

func TestPortsOutliveTheirService(t *testing.T) {
	node := newNode(t, iox2.ServiceTypeLocal)
	service, err := node.PublishSubscribe(serviceName(t)).OpenOrCreate()
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	publisher, err := service.PublisherBuilder().Create()
	if err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	defer publisher.Close()
	subscriber, err := service.SubscriberBuilder().Create()
	if err != nil {
		t.Fatalf("create subscriber: %v", err)
	}
	defer subscriber.Close()
	service.Close()

	if _, err := publisher.Send([]byte("still there")); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got := receiveBytes(t, subscriber); string(got) != "still there" {
		t.Errorf("received %q, want %q", got, "still there")
	}
}

func TestPublishAndReceiveConcurrently(t *testing.T) {
	const senders, messagesPerSender = 4, 500
	for _, serviceType := range serviceTypes {
		t.Run(serviceType.String(), func(t *testing.T) {
			publisher, subscriber := ports(t, newNode(t, serviceType), iox2.FixedSizePayload[reading](readingTypeName))

			// Several goroutines share one publisher. The subscriber buffer is
			// small and overflows safely, so samples get lost when the receiver
			// is slower. What arrives has to be intact and in order per sender.
			var wg sync.WaitGroup
			for sender := range senders {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := range messagesPerSender {
						r := reading{ID: uint32(sender<<16 | i), Flags: uint8(sender), Value: float64(i)}
						if _, err := iox2.SendValue(publisher, r); err != nil {
							t.Errorf("SendValue: %v", err)
							return
						}
					}
				}()
			}
			sent := make(chan struct{})
			go func() {
				wg.Wait()
				close(sent)
			}()

			lastIndex := make(map[uint8]int)
			received := 0
			for done := false; !done; {
				select {
				case <-sent:
					done = true // drain what is left
				default:
				}
				for {
					r, ok, err := iox2.ReceiveValue[reading](subscriber)
					if err != nil {
						t.Fatalf("ReceiveValue: %v", err)
					}
					if !ok {
						break
					}
					received++
					index := int(r.ID & 0xFFFF)
					if r.ID>>16 != uint32(r.Flags) || r.Value != float64(index) {
						t.Fatalf("received a torn sample: %+v", r)
					}
					if last, seen := lastIndex[r.Flags]; seen && index <= last {
						t.Fatalf("sender %d: sample %d arrived after sample %d", r.Flags, index, last)
					}
					lastIndex[r.Flags] = index
				}
			}
			if received == 0 {
				t.Error("received no sample at all")
			}
			t.Logf("received %d of %d samples", received, senders*messagesPerSender)
		})
	}
}

func TestCloseIsSafeToRaceWithUse(t *testing.T) {
	publisher, subscriber := ports(t, newNode(t, iox2.ServiceTypeLocal), iox2.BytesPayload())

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(3)
		go func() {
			defer wg.Done()
			for range 100 {
				if _, err := publisher.Send([]byte("x")); err != nil && !errors.Is(err, iox2.ErrClosed) {
					t.Errorf("Send: %v", err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			for range 100 {
				// Several goroutines hold samples at once, more than the
				// subscriber hands out.
				sample, ok, err := subscriber.Receive()
				if err != nil && !errors.Is(err, iox2.ErrClosed) && !errors.Is(err, iox2.ReceiveErrorExceedsMaxBorrows) {
					t.Errorf("Receive: %v", err)
				}
				if ok {
					sample.Payload()
					sample.Close()
				}
			}
		}()
		go func() {
			defer wg.Done()
			publisher.Close()
			subscriber.Close()
		}()
	}
	wg.Wait()
}

func ExamplePublisher_Send() {
	node, err := iox2.NewNodeBuilder().Create(iox2.ServiceTypeLocal)
	if err != nil {
		panic(err)
	}
	defer node.Close()
	service, err := node.PublishSubscribe("Example/Greeting").OpenOrCreate()
	if err != nil {
		panic(err)
	}
	defer service.Close()
	publisher, err := service.PublisherBuilder().Create()
	if err != nil {
		panic(err)
	}
	defer publisher.Close()
	subscriber, err := service.SubscriberBuilder().Create()
	if err != nil {
		panic(err)
	}
	defer subscriber.Close()

	if _, err := publisher.Send([]byte("hello")); err != nil {
		panic(err)
	}
	sample, ok, err := subscriber.Receive()
	if err != nil || !ok {
		panic("no sample")
	}
	defer sample.Close()
	fmt.Printf("%s\n", sample.Payload())
	// Output: hello
}

/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package bus

import (
    "context"
    "github.com/golang/protobuf/proto"
    "github.com/nalej/derrors"
    "github.com/nalej/nalej-bus/pkg/bus"
    "github.com/nalej/nalej-bus/pkg/queue/application/ops"
)

// Structures and operators designed to manipulate the queue operations for the application ops queue.

type BusManager struct {
    producer *ops.ApplicationOpsProducer
}

// Create a new bus manager
// params:
//  client the implementation of the queuing protocol
//  name of the producer to be generated
// return:
//  bus manager instance
//  error if any
func NewBusManager(client bus.NalejClient, name string) (*BusManager, derrors.Error) {
    producer, err := ops.NewApplicationOpsProducer(client, name)
    if err != nil {
        return nil, err
    }

    return &BusManager{producer: producer}, nil
}

// Send messages to the queue. If the sent proto message is not allowed by the queue and error will be triggered.
func (b BusManager) Send(ctx context.Context, msg proto.Message) derrors.Error {
    return b.producer.Send(ctx, msg)
}

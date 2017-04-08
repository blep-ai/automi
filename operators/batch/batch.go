package batch

import (
	"context"
	"fmt"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

// BatchOperator is an executor that batches incoming streamed items based
// on provided criteria.  The batched items are streamed on the
// ouptut channel for downstream processing.
type BatchOperator struct {
	ctx    context.Context
	input  <-chan interface{}
	output chan interface{}
	log    *log.Logger
	size   int
}

// New returns a new BatchOperator operator
func New(ctx context.Context) *BatchOperator {
	log := autoctx.GetLogger(ctx)
	op := new(BatchOperator)
	op.ctx = ctx
	op.log = log
	op.output = make(chan interface{}, 1024)
	op.size = 1024 * 10
	return op
}

// SetInput sets the input channel for the executor node
func (op *BatchOperator) SetInput(in <-chan interface{}) {
	op.input = in
}

// GetOutput returns the output channel of the executer node
func (op *BatchOperator) GetOutput() <-chan interface{} {
	return op.output
}

// Exec is the execution starting point for the executor node.
func (op *BatchOperator) Exec() (err error) {
	if op.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	go func() {
		defer func() {
			close(op.output)
			op.log.Print("component shutting down")
		}()
		batch := make([]interface{}, 0, op.size)
		counter := 0
		for {
			select {
			case item, opened := <-op.input:
				if !opened {
					return
				}
				batch = append(batch, item)
				if counter < op.size-1 {
					counter++
					continue
				}
				if counter >= op.size-1 {
					op.output <- batch
					counter = 0
					batch = make([]interface{}, 0, op.size)
				}
			}
		}
	}()
	return nil
}
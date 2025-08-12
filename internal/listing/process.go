package listing

import (
	"context"
	"sync"
)

const channelSize = 10000

type Item any

type Processor func(Item) error

type WorkerFunc[S any] func(ctx context.Context, input S, output chan<- Item, errChan chan<- error)

func Process[S any](
	ctx context.Context, inputs []S, workerFunc WorkerFunc[S], processResult Processor) error {
	if len(inputs) == 0 {
		return nil
	}

	resultChan := make(chan Item, channelSize)
	errChan := make(chan error, len(inputs))
	var wg sync.WaitGroup

	for _, input := range inputs {
		wg.Add(1)
		go func(input S) {
			defer wg.Done()
			workerFunc(ctx, input, resultChan, errChan)
		}(input)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				select {
				case err := <-errChan:
					return err
				default:
					return nil
				}
			}
			if err := processResult(result); err != nil {
				return err
			}

		case err := <-errChan:
			return err

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

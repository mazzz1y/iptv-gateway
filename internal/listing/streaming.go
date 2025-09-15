package listing

import (
	"context"
	"io"
	"sync"
	"time"
)

const (
	bufferTicker    = time.Second
	bufferBatchSize = 1000
	bufferSize      = 100000
)

type initFunc func(ctx context.Context, url string) (Decoder, io.ReadCloser, error)

type BaseDecoder struct {
	decoder      Decoder
	reader       io.ReadCloser
	err          error
	itemBuffer   []any
	bufferMu     sync.Mutex
	bufferCtx    context.Context
	cancelBuffer context.CancelFunc
	url          string
	initFunc     initFunc
	bufferWG     sync.WaitGroup
}

func NewLazyBaseDecoder(url string, init initFunc) *BaseDecoder {
	return &BaseDecoder{
		url:        url,
		initFunc:   init,
		itemBuffer: make([]any, 0, bufferSize),
	}
}

func (d *BaseDecoder) NextItem() (any, error) {
	if d.err != nil {
		return nil, d.err
	}

	d.bufferMu.Lock()
	if len(d.itemBuffer) > 0 {
		item := d.itemBuffer[0]
		d.itemBuffer = d.itemBuffer[1:]
		d.bufferMu.Unlock()
		return item, nil
	}
	d.bufferMu.Unlock()

	item, err := d.decoder.Decode()
	if err == io.EOF {
		d.drainReader()
	}

	return item, err
}

func (d *BaseDecoder) StartBuffering(ctx context.Context) error {
	if err := d.init(ctx); err != nil {
		d.err = err
		return err
	}

	d.bufferCtx, d.cancelBuffer = context.WithCancel(ctx)
	d.bufferWG.Add(1)

	go func() {
		ticker := time.NewTicker(bufferTicker)
		defer ticker.Stop()
		defer d.bufferWG.Done()

		batch := make([]any, 0, bufferBatchSize)

		for {
			select {
			case <-d.bufferCtx.Done():
				return
			case <-ticker.C:
				batch = batch[:0]
				for i := 0; i < bufferBatchSize; i++ {
					item, err := d.decoder.Decode()
					if err == io.EOF {
						if len(batch) > 0 {
							d.AddToBuffer(batch)
						}
						return
					}
					if err != nil {
						d.err = err
						return
					}
					batch = append(batch, item)
				}
				d.AddToBuffer(batch)
			}
		}
	}()

	return nil
}

func (d *BaseDecoder) AddToBuffer(items ...any) {
	d.bufferMu.Lock()
	d.itemBuffer = append(d.itemBuffer, items...)
	d.bufferMu.Unlock()
}

func (d *BaseDecoder) StopBuffer() {
	if d.cancelBuffer != nil {
		d.cancelBuffer()
	}
	d.bufferWG.Wait()
}

func (d *BaseDecoder) Close() error {
	if d.cancelBuffer != nil {
		d.cancelBuffer()
	}
	if d.reader != nil {
		return d.reader.Close()
	}
	return nil
}

func (d *BaseDecoder) init(ctx context.Context) error {
	if d.decoder != nil || d.reader != nil {
		return nil
	}

	decoder, reader, err := d.initFunc(ctx, d.url)
	if err != nil {
		return err
	}

	d.decoder = decoder
	d.reader = reader

	return nil
}

func (d *BaseDecoder) drainReader() {
	if d.reader == nil {
		return
	}

	buf := make([]byte, 1024)
	for {
		_, err := d.reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}
}

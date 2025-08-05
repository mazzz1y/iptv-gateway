package common

import (
	"context"
	"io"
	"iptv-gateway/internal/cache"
	"iptv-gateway/internal/manager"
)

type HTTPClient interface {
	NewReader(ctx context.Context, url string) (*cache.Reader, error)
}

type Decoder interface {
	Decode() (any, error)
	Close() error
}

type DecoderFactory interface {
	NewDecoder(reader *cache.Reader) Decoder
}

type BaseStreamer struct {
	Subscriptions        []*manager.Subscription
	HTTPClient           HTTPClient
	CurrentDecoder       Decoder
	CurrentSubscription  *manager.Subscription
	PendingSubscriptions []*manager.Subscription
	PendingReaders       []*cache.Reader
	DecoderFactory       DecoderFactory
}

func NewBaseStreamer(subs []*manager.Subscription, httpClient HTTPClient, decoderFactory DecoderFactory) *BaseStreamer {
	return &BaseStreamer{
		Subscriptions:        subs,
		HTTPClient:           httpClient,
		PendingSubscriptions: subs,
		PendingReaders:       []*cache.Reader{},
		DecoderFactory:       decoderFactory,
	}
}

func (s *BaseStreamer) Close() error {
	var err error

	if s.CurrentDecoder != nil {
		err = s.CurrentDecoder.Close()
		s.CurrentDecoder = nil
	}

	for _, reader := range s.PendingReaders {
		if closeErr := reader.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	s.PendingReaders = nil
	return err
}

func (s *BaseStreamer) NextItem(ctx context.Context, getSourcesFunc func(*manager.Subscription) []string) (any, error) {
	if s.CurrentDecoder == nil && len(s.PendingSubscriptions) == 0 && len(s.PendingReaders) == 0 {
		return nil, io.EOF
	}

	for {
		if s.CurrentDecoder != nil {
			item, err := s.CurrentDecoder.Decode()
			if err == io.EOF {
				if s.CurrentDecoder != nil {
					s.CurrentDecoder.Close()
					s.CurrentDecoder = nil
				}
				if len(s.PendingReaders) > 0 {
					s.CurrentDecoder = s.DecoderFactory.NewDecoder(s.PendingReaders[0])
					s.PendingReaders = s.PendingReaders[1:]
					continue
				}
				continue
			}
			if err != nil {
				s.Close()
				return nil, err
			}
			return item, nil
		}

		if len(s.PendingSubscriptions) == 0 {
			return nil, io.EOF
		}

		s.Close()

		nextSubscription := s.PendingSubscriptions[0]
		s.PendingSubscriptions = s.PendingSubscriptions[1:]
		s.CurrentSubscription = nextSubscription

		var readers []*cache.Reader

		for _, source := range getSourcesFunc(nextSubscription) {
			r, err := s.HTTPClient.NewReader(ctx, source)
			if err != nil {
				return nil, err
			}
			readers = append(readers, r)
		}

		if len(readers) > 0 {
			s.CurrentDecoder = s.DecoderFactory.NewDecoder(readers[0])
			s.PendingReaders = readers[1:]
		}
	}
}

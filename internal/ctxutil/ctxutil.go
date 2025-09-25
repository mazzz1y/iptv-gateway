package ctxutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type contextKey string

const (
	clientKey        contextKey = "client"
	clientNameKey    contextKey = "client_name"
	requestIDKey     contextKey = "request_id"
	channelHiddenKey contextKey = "channel_hidden"
	streamDataKey    contextKey = "stream_data"
	providerKey      contextKey = "provider"
	providerTypeKey  contextKey = "provider_type"
	providerNameKey  contextKey = "provider_name"
	streamIDKey      contextKey = "stream_id"
	channelIDKey     contextKey = "channel_id"
	semaphoreNameKey contextKey = "semaphore_name"
)

type Listing interface {
	Name() string
	Type() string
}

func WithRequestID(ctx context.Context) context.Context {
	b := make([]byte, 4)
	rand.Read(b)
	return context.WithValue(ctx, requestIDKey, hex.EncodeToString(b))
}

func WithClient(ctx context.Context, client any) context.Context {
	if namer, ok := client.(interface{ Name() string }); ok {
		ctx = context.WithValue(ctx, clientNameKey, namer.Name())
	}
	return context.WithValue(ctx, clientKey, client)
}

func WithProvider(ctx context.Context, src Listing) context.Context {
	ctx = context.WithValue(ctx, providerNameKey, src.Name())
	ctx = context.WithValue(ctx, providerTypeKey, src.Type())
	return context.WithValue(ctx, providerKey, src)
}

func WithStreamData(ctx context.Context, data any) context.Context {
	return context.WithValue(ctx, streamDataKey, data)
}

func WithStreamID(ctx context.Context, streamID string) context.Context {
	return context.WithValue(ctx, streamIDKey, streamID)
}

func WithChannelName(ctx context.Context, channelID string) context.Context {
	return context.WithValue(ctx, channelIDKey, channelID)
}

func WithChannelHidden(ctx context.Context, hidden bool) context.Context {
	return context.WithValue(ctx, channelHiddenKey, hidden)
}

func WithSemaphoreName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, semaphoreNameKey, name)
}

func WithRequestType(ctx context.Context, requestType string) context.Context {
	return context.WithValue(ctx, channelHiddenKey, requestType)
}

func RequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		return v.(string)
	}
	return ""
}

func Client(ctx context.Context) any {
	return ctx.Value(clientKey)
}

func ClientName(ctx context.Context) string {
	if v := ctx.Value(clientNameKey); v != nil {
		return v.(string)
	}
	return ""
}

func Provider(ctx context.Context) any {
	return ctx.Value(providerKey)
}

func ProviderName(ctx context.Context) string {
	if v := ctx.Value(providerNameKey); v != nil {
		return v.(string)
	}
	return ""
}

func ProviderType(ctx context.Context) string {
	if v := ctx.Value(providerTypeKey); v != nil {
		return v.(string)
	}
	return ""
}

func StreamData(ctx context.Context) any {
	return ctx.Value(streamDataKey)
}

func StreamID(ctx context.Context) string {
	if v := ctx.Value(streamIDKey); v != nil {
		return v.(string)
	}
	return ""
}

func ChannelID(ctx context.Context) string {
	if v := ctx.Value(channelIDKey); v != nil {
		return v.(string)
	}
	return ""
}

func ChannelHidden(ctx context.Context) bool {
	if v := ctx.Value(channelHiddenKey); v != nil {
		return v.(bool)
	}
	return false
}

func SemaphoreName(ctx context.Context) string {
	if v := ctx.Value(semaphoreNameKey); v != nil {
		return v.(string)
	}
	return ""
}

func RequestType(ctx context.Context) string {
	if reqType, ok := ctx.Value(channelHiddenKey).(string); ok {
		return reqType
	}
	return "unknown"
}

func LogFields(ctx context.Context) []any {
	fields := make([]any, 0, 16)

	if id := RequestID(ctx); id != "" {
		fields = append(fields, "request_id", id)
	}
	if name := ClientName(ctx); name != "" {
		fields = append(fields, "client_name", name)
	}
	if prName := ProviderName(ctx); prName != "" {
		fields = append(fields, "provider_name", prName)
	}
	if pr := ProviderType(ctx); pr != "" {
		fields = append(fields, "provider_type", pr)
	}
	if id := StreamID(ctx); id != "" {
		fields = append(fields, "stream_id", id)
	}
	if name := SemaphoreName(ctx); name != "" {
		fields = append(fields, "semaphore_name", name)
	}
	if id := ChannelID(ctx); id != "" {
		fields = append(fields, "channel_id", id)
	}

	return fields
}

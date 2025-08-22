package ctxutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type contextKey string

const (
	clientKey           contextKey = "client"
	clientNameKey       contextKey = "client_name"
	requestIDKey        contextKey = "request_id"
	requestTypeKey      contextKey = "request_type"
	streamDataKey       contextKey = "stream_data"
	subscriptionKey     contextKey = "subscription"
	subscriptionNameKey contextKey = "subscription_name"
	streamIDKey         contextKey = "stream_id"
	channelIDKey        contextKey = "channel_id"
	semaphoreNameKey    contextKey = "semaphore_name"
)

func WithRequestID(ctx context.Context) context.Context {
	b := make([]byte, 4)
	rand.Read(b)
	return context.WithValue(ctx, requestIDKey, hex.EncodeToString(b))
}

func WithClient(ctx context.Context, client any) context.Context {
	if namer, ok := client.(interface{ GetName() string }); ok {
		ctx = context.WithValue(ctx, clientNameKey, namer.GetName())
	}
	return context.WithValue(ctx, clientKey, client)
}

func WithSubscription(ctx context.Context, sub any) context.Context {
	if namer, ok := sub.(interface{ GetName() string }); ok {
		ctx = context.WithValue(ctx, subscriptionNameKey, namer.GetName())
	}
	return context.WithValue(ctx, subscriptionKey, sub)
}

func WithStreamData(ctx context.Context, data any) context.Context {
	return context.WithValue(ctx, streamDataKey, data)
}

func WithStreamID(ctx context.Context, streamID string) context.Context {
	return context.WithValue(ctx, streamIDKey, streamID)
}

func WithChannelID(ctx context.Context, channelID string) context.Context {
	return context.WithValue(ctx, channelIDKey, channelID)
}

func WithSemaphoreName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, semaphoreNameKey, name)
}

func WithRequestType(ctx context.Context, requestType string) context.Context {
	return context.WithValue(ctx, requestTypeKey, requestType)
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

func Subscription(ctx context.Context) any {
	return ctx.Value(subscriptionKey)
}

func SubscriptionName(ctx context.Context) string {
	if v := ctx.Value(subscriptionNameKey); v != nil {
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

func SemaphoreName(ctx context.Context) string {
	if v := ctx.Value(semaphoreNameKey); v != nil {
		return v.(string)
	}
	return ""
}

func RequestType(ctx context.Context) string {
	if reqType, ok := ctx.Value(requestTypeKey).(string); ok {
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
	if name := SubscriptionName(ctx); name != "" {
		fields = append(fields, "subscription_name", name)
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

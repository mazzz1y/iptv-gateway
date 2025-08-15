package xmltv

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"iptv-gateway/internal/cache"
	"iptv-gateway/internal/client"
	"iptv-gateway/internal/config"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
)

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Close() {}

func (m *MockHTTPClient) NewReader(ctx context.Context, url string) (*cache.Reader, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*cache.Reader), args.Error(1)
}

func createMockReader(readCloser io.ReadCloser, contentType string) *cache.Reader {
	return &cache.Reader{
		URL:        "test://example.com",
		Name:       "test",
		FilePath:   "test.gz",
		MetaPath:   "test.meta",
		ReadCloser: readCloser,
	}
}

type MockDecoder struct {
	mock.Mock
	closed bool
}

func (m *MockDecoder) Decode() (interface{}, error) {
	return nil, io.EOF
}

func (m *MockDecoder) Close() error {
	m.closed = true
	return nil
}

func createTestSubscription(name string, epgs []string) (*client.Subscription, error) {
	sem := semaphore.NewWeighted(1)
	return client.NewSubscription(
		name,
		nil,
		nil,
		epgs,
		config.Proxy{},
		[]config.RuleAction{},
		sem,
	)
}

func TestNewStreamer(t *testing.T) {
	var subscriptions []*client.Subscription
	httpClient := &MockHTTPClient{}
	channels := map[string]bool{"channel1": true}

	streamer := NewStreamer(subscriptions, httpClient, channels)
	assert.NotNil(t, streamer)
	assert.Equal(t, channels, streamer.channels)
	assert.NotNil(t, streamer.addedChannels)
	assert.NotNil(t, streamer.addedProgrammes)
}

func TestStreamer_WriteTo(t *testing.T) {
	ctx := context.Background()
	streamer := NewStreamer([]*client.Subscription{}, &MockHTTPClient{}, nil)
	buf := bytes.NewBuffer(nil)
	_, err := streamer.WriteTo(ctx, buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no EPG sources found")

	httpClient := new(MockHTTPClient)

	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <channel id="channel1">
	<display-name>Channel 1</display-name>
	<icon src="http://example.com/icon.png" width="100" height="100"/>
  </channel>
  <programme start="20230101120000 +0000" channel="channel1">
	<title>Test Programme</title>
	<desc>Programme description</desc>
	<icon src="http://example.com/prog.png" width="100" height="100"/>
  </programme>
</tv>`

	sub, err := createTestSubscription("test-subscription", []string{"http://example.com/epg.xml"})
	require.NoError(t, err)

	httpClient.On("NewReader", mock.Anything, "http://example.com/epg.xml").Return(
		createMockReader(io.NopCloser(strings.NewReader(xmlContent)), ""),
		nil,
	)

	channels := map[string]bool{"channel1": true}
	streamer = NewStreamer([]*client.Subscription{sub}, httpClient, channels)
	buf = bytes.NewBuffer(nil)
	_, err = streamer.WriteTo(ctx, buf)
	require.NoError(t, err)
	result := buf.String()
	assert.NotEmpty(t, result, "Expected non-empty XML output")
	assert.Contains(t, strings.ToLower(result), "<channel id=\"channel1\">")
	assert.Contains(t, strings.ToLower(result), "<programme start=\"")
	assert.Contains(t, result, "http://example.com/icon.png")
}

func TestStreamer_WriteToGzip(t *testing.T) {
	ctx := context.Background()
	httpClient := new(MockHTTPClient)

	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <channel id="channel1">
	<display-name>Channel 1</display-name>
  </channel>
</tv>`

	sub, err := createTestSubscription("test-subscription", []string{"http://example.com/epg.xml"})
	require.NoError(t, err)

	httpClient.On("NewReader", mock.Anything, "http://example.com/epg.xml").Return(
		createMockReader(io.NopCloser(strings.NewReader(xmlContent)), ""),
		nil,
	)

	channels := map[string]bool{"channel1": true}
	streamer := NewStreamer([]*client.Subscription{sub}, httpClient, channels)
	buf := bytes.NewBuffer(nil)
	_, err = streamer.WriteToGzip(ctx, buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes(), "Expected non-empty gzip output")

	gzipReader, err := gzip.NewReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	uncompressed, err := io.ReadAll(gzipReader)
	require.NoError(t, err)

	result := string(uncompressed)
	assert.NotEmpty(t, result, "Expected non-empty XML output after decompression")
	assert.Contains(t, strings.ToLower(result), "<channel id=\"channel1\">")
}

func TestStreamerWithMultipleEPGSources(t *testing.T) {
	ctx := context.Background()
	httpClient := new(MockHTTPClient)

	xmlContent1 := `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <channel id="channel1">
	<display-name>Channel 1</display-name>
	<icon src="http://example.com/icon1.png"/>
  </channel>
  <programme start="20230101120000 +0000" channel="channel1">
	<title>Morning Show</title>
	<desc>A morning program</desc>
  </programme>
</tv>`

	xmlContent2 := `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <channel id="channel2">
	<display-name>Channel 2</display-name>
	<icon src="http://example.com/icon2.png"/>
  </channel>
  <programme start="20230101140000 +0000" channel="channel1">
	<title>Afternoon Show</title>
	<desc>An afternoon program</desc>
  </programme>
</tv>`

	mockReader1 := createMockReader(io.NopCloser(strings.NewReader(xmlContent1)), "")
	mockReader2 := createMockReader(io.NopCloser(strings.NewReader(xmlContent2)), "")

	httpClient.On("NewReader", mock.Anything, "http://example.com/epg1.xml").Return(mockReader1, nil)
	httpClient.On("NewReader", mock.Anything, "http://example.com/epg2.xml").Return(mockReader2, nil)

	sub, err := createTestSubscription(
		"test-subscription",
		[]string{
			"http://example.com/epg1.xml",
			"http://example.com/epg2.xml",
		},
	)
	require.NoError(t, err)

	channels := map[string]bool{
		"channel1": true,
		"channel2": true,
	}

	streamer := NewStreamer([]*client.Subscription{sub}, httpClient, channels)

	buffer := &bytes.Buffer{}

	n, err := streamer.WriteTo(ctx, buffer)
	require.NoError(t, err)
	require.Greater(t, n, int64(0))

	output := buffer.String()

	assert.Contains(t, output, "<channel id=\"channel1\">")
	assert.Contains(t, output, "<title>Morning Show</title>")
	assert.Contains(t, output, "http://example.com/icon1.png")

	assert.Contains(t, output, "<channel id=\"channel2\">")
	assert.Contains(t, output, "<title>Afternoon Show</title>")
	assert.Contains(t, output, "http://example.com/icon2.png")

	httpClient.AssertExpectations(t)
}

func TestStreamerWithMultipleSubscriptionsAndEPGs(t *testing.T) {
	ctx := context.Background()
	httpClient := new(MockHTTPClient)

	xmlContent1 := `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <channel id="news1">
	<display-name>News Channel</display-name>
  </channel>
</tv>`

	xmlContent2 := `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <channel id="movies1">
	<display-name>Movies Channel</display-name>
  </channel>
</tv>`

	mockReader1 := createMockReader(io.NopCloser(strings.NewReader(xmlContent1)), "")
	mockReader2 := createMockReader(io.NopCloser(strings.NewReader(xmlContent2)), "")

	httpClient.On("NewReader", mock.Anything, "http://example.com/sub1_epg.xml").Return(mockReader1, nil)
	httpClient.On("NewReader", mock.Anything, "http://example.com/sub2_epg.xml").Return(mockReader2, nil)

	sub1, err := createTestSubscription(
		"subscription-1",
		[]string{"http://example.com/sub1_epg.xml"},
	)
	require.NoError(t, err)

	sub2, err := createTestSubscription(
		"subscription-2",
		[]string{"http://example.com/sub2_epg.xml"},
	)
	require.NoError(t, err)

	channels := map[string]bool{
		"news1":   true,
		"movies1": true,
	}

	streamer := NewStreamer([]*client.Subscription{sub1, sub2}, httpClient, channels)

	buffer := &bytes.Buffer{}

	_, err = streamer.WriteTo(ctx, buffer)
	require.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "<channel id=\"news1\">")
	assert.Contains(t, output, "<channel id=\"movies1\">")

	httpClient.AssertExpectations(t)
}

func TestStreamerEmptyEPGSubscription(t *testing.T) {
	ctx := context.Background()
	httpClient := new(MockHTTPClient)

	emptySub, err := createTestSubscription(
		"empty-subscription",
		[]string{},
	)
	require.NoError(t, err)

	validSub, err := createTestSubscription(
		"valid-subscription",
		[]string{"http://example.com/epg.xml"},
	)
	require.NoError(t, err)

	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <channel id="test1">
	<display-name>Test Channel</display-name>
  </channel>
</tv>`

	mockReader := createMockReader(io.NopCloser(strings.NewReader(xmlContent)), "")
	httpClient.On("NewReader", mock.Anything, "http://example.com/epg.xml").Return(mockReader, nil)

	channels := map[string]bool{"test1": true}

	streamer := NewStreamer([]*client.Subscription{emptySub, validSub}, httpClient, channels)

	buffer := &bytes.Buffer{}
	_, err = streamer.WriteTo(ctx, buffer)
	require.NoError(t, err)

	assert.Contains(t, buffer.String(), "<channel id=\"test1\">")

	httpClient.AssertExpectations(t)
}

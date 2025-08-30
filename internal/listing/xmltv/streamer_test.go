package xmltv

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"iptv-gateway/internal/app"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/urlgen"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
)

func createStreamer(subscriptions []listing.Subscription, httpClient listing.HTTPClient, channels map[string]bool) *Streamer {
	return &Streamer{
		subscriptions:   subscriptions,
		httpClient:      httpClient,
		channels:        channels,
		addedChannels:   make(map[string]bool, DefaultChannelMapSize),
		addedProgrammes: make(map[string]bool, DefaultProgrammeMapSize),
	}
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func createTestSubscription(name string, epgs []string) (*app.Subscription, error) {
	sem := semaphore.NewWeighted(1)
	generator, err := urlgen.NewGenerator("http://localhost", "secret")
	if err != nil {
		return nil, err
	}
	return app.NewSubscription(
		name,
		*generator,
		nil,
		epgs,
		config.Proxy{},
		[]rules.RuleAction{},
		sem,
	)
}

func TestNewStreamer(t *testing.T) {
	var subscriptions []listing.Subscription
	httpClient := &MockHTTPClient{}
	channels := map[string]bool{"channel1": true}

	streamer := createStreamer(subscriptions, httpClient, channels)
	assert.NotNil(t, streamer)
	assert.Equal(t, channels, streamer.channels)
	assert.NotNil(t, streamer.addedChannels)
	assert.NotNil(t, streamer.addedProgrammes)
}

func TestStreamer_WriteTo(t *testing.T) {
	ctx := context.Background()
	streamer := createStreamer([]listing.Subscription{}, &MockHTTPClient{}, nil)
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

	response := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(xmlContent)),
	}

	httpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" && req.URL.String() == "http://example.com/epg.xml"
	})).Return(response, nil)

	channels := map[string]bool{"channel1": true}
	streamer = createStreamer([]listing.Subscription{sub}, httpClient, channels)
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

	response := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(xmlContent)),
	}

	httpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" && req.URL.String() == "http://example.com/epg.xml"
	})).Return(response, nil)

	channels := map[string]bool{"channel1": true}
	streamer := createStreamer([]listing.Subscription{sub}, httpClient, channels)
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

	response1 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(xmlContent1)),
	}
	response2 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(xmlContent2)),
	}

	httpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" && req.URL.String() == "http://example.com/epg1.xml"
	})).Return(response1, nil)

	httpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" && req.URL.String() == "http://example.com/epg2.xml"
	})).Return(response2, nil)

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

	streamer := createStreamer([]listing.Subscription{sub}, httpClient, channels)

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

	response1 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(xmlContent1)),
	}
	response2 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(xmlContent2)),
	}

	httpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" && req.URL.String() == "http://example.com/sub1_epg.xml"
	})).Return(response1, nil)

	httpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" && req.URL.String() == "http://example.com/sub2_epg.xml"
	})).Return(response2, nil)

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

	streamer := createStreamer([]listing.Subscription{sub1, sub2}, httpClient, channels)

	buffer := &bytes.Buffer{}

	_, err = streamer.WriteTo(ctx, buffer)
	require.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "<channel id=\"news1\">")
	assert.Contains(t, output, "<channel id=\"movies1\">")

	httpClient.AssertExpectations(t)
}

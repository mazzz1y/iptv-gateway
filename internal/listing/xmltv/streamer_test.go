package xmltv

import (
	"bytes"
	"context"
	"io"
	"iptv-gateway/internal/app"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/urlgen"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createStreamer(subscriptions []listing.EPGSubscription, httpClient listing.HTTPClient, channelIDToName map[string]string) *Streamer {
	channelLen := len(channelIDToName)
	approxProgrammeLen := 300 * channelLen

	return &Streamer{
		subscriptions:    subscriptions,
		httpClient:       httpClient,
		channelIDToName:  channelIDToName,
		addedChannelIDs:  make(map[string]bool, channelLen),
		addedProgrammes:  make(map[string]bool, approxProgrammeLen),
		channelIDMapping: make(map[string]string, channelLen),
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

func createTestSubscription(name string, epgs []string) (*app.EPGSubscription, error) {
	generator, err := urlgen.NewGenerator("http://localhost", "secret")
	if err != nil {
		return nil, err
	}
	return app.NewEPGSubscription(
		name,
		*generator,
		epgs,
		config.Proxy{},
	)
}

func TestNewStreamer(t *testing.T) {
	var subscriptions []listing.EPGSubscription
	httpClient := &MockHTTPClient{}
	channels := map[string]string{"channel1": "Channel One"}

	streamer := createStreamer(subscriptions, httpClient, channels)
	assert.NotNil(t, streamer)
	assert.Equal(t, channels, streamer.channelIDToName)
	assert.NotNil(t, streamer.addedChannelIDs)
	assert.NotNil(t, streamer.addedProgrammes)
}

func TestStreamer_WriteTo(t *testing.T) {
	ctx := context.Background()
	streamer := createStreamer([]listing.EPGSubscription{}, &MockHTTPClient{}, nil)
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

	channels := map[string]string{"channel1": "Channel One"}
	streamer = createStreamer([]listing.EPGSubscription{sub}, httpClient, channels)
	buf = bytes.NewBuffer(nil)
	_, err = streamer.WriteTo(ctx, buf)
	require.NoError(t, err)
	result := buf.String()
	assert.NotEmpty(t, result, "Expected non-empty XML output")
	assert.Contains(t, strings.ToLower(result), "<channel id=\"channel1\">")
	assert.Contains(t, strings.ToLower(result), "<programme start=\"")
	assert.Contains(t, result, "http://localhost/")
	assert.Contains(t, result, "/f.png")
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

	channels := map[string]string{
		"channel1": "Channel One",
		"channel2": "Channel Two",
	}

	streamer := createStreamer([]listing.EPGSubscription{sub}, httpClient, channels)

	buffer := &bytes.Buffer{}

	n, err := streamer.WriteTo(ctx, buffer)
	require.NoError(t, err)
	require.Greater(t, n, int64(0))

	output := buffer.String()

	assert.Contains(t, output, "<channel id=\"channel1\">")
	assert.Contains(t, output, "<title>Morning Show</title>")
	assert.Contains(t, output, "http://localhost/")
	assert.Contains(t, output, "<channel id=\"channel2\">")
	assert.Contains(t, output, "<title>Afternoon Show</title>")

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

	channels := map[string]string{
		"news1":   "News Channel",
		"movies1": "Movies Channel",
	}

	streamer := createStreamer([]listing.EPGSubscription{sub1, sub2}, httpClient, channels)

	buffer := &bytes.Buffer{}

	_, err = streamer.WriteTo(ctx, buffer)
	require.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "<channel id=\"news1\">")
	assert.Contains(t, output, "<channel id=\"movies1\">")

	httpClient.AssertExpectations(t)
}

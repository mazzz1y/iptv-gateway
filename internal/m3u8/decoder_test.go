package m3u8

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoderBasicParsing(t *testing.T) {
	sampleM3U := `#EXTM3U
#EXTINF:-1 tvg-id="test1" tvg-name="Test Channel 1" tvg-logo="http://example.com/logo.png" group-title="News", Test Channel 1
http://example.com/stream1
#EXTINF:0 tvg-id="test2" tvg-name="Test Channel 2", Test Channel 2
http://example.com/stream2`

	decoder := NewDecoder(strings.NewReader(sampleM3U))

	item, err := decoder.Decode()
	require.NoError(t, err)
	track, ok := item.(*Track)
	require.True(t, ok, "Expected a Track object")

	assert.Equal(t, "Test Channel 1", track.Name)
	assert.Equal(t, -1.0, track.Length)
	assert.Equal(t, "http://example.com/stream1", track.URI.String())
	assert.Equal(t, "test1", track.Attrs[AttrTvgID])
	assert.Equal(t, "Test Channel 1", track.Attrs[AttrTvgName])
	assert.Equal(t, "http://example.com/logo.png", track.Attrs[AttrTvgLogo])
	assert.Equal(t, "News", track.Attrs[AttrGroupTitle])

	item, err = decoder.Decode()
	require.NoError(t, err)
	track, ok = item.(*Track)
	require.True(t, ok, "Expected a Track object")

	assert.Equal(t, "Test Channel 2", track.Name)
	assert.Equal(t, 0.0, track.Length)
	assert.Equal(t, "http://example.com/stream2", track.URI.String())
	assert.Equal(t, "test2", track.Attrs[AttrTvgID])
	assert.Equal(t, "Test Channel 2", track.Attrs[AttrTvgName])

	_, err = decoder.Decode()
	assert.Equal(t, io.EOF, err)
}

func TestDecoderWithTags(t *testing.T) {
	sampleM3U := `#EXTM3U
#EXTINF:-1 tvg-id="test1", Channel with Tags
#EXTVLCOPT:http-user-agent=Mozilla/5.0
#KODIPROP:inputstream=inputstream.adaptive
http://example.com/stream1`

	decoder := NewDecoder(strings.NewReader(sampleM3U))

	item, err := decoder.Decode()
	require.NoError(t, err)
	track, ok := item.(*Track)
	require.True(t, ok, "Expected a Track object")

	assert.Equal(t, "Channel with Tags", track.Name)
	assert.Equal(t, "test1", track.Attrs[AttrTvgID])
	assert.Equal(t, "http://example.com/stream1", track.URI.String())

	assert.NotEmpty(t, track.Tags)
	assert.Contains(t, track.Tags, "EXTVLCOPT")
}

func TestDecoderInvalidFormat(t *testing.T) {
	invalidM3U := `#EXTINF:-1, Test Channel
http://example.com/stream`
	decoder := NewDecoder(strings.NewReader(invalidM3U))
	_, err := decoder.Decode()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing #EXTM3U header")
}

func TestDecoderEmptyFile(t *testing.T) {
	decoder := NewDecoder(strings.NewReader(""))
	_, err := decoder.Decode()
	assert.Equal(t, io.EOF, err)
}

func TestDecoderMissingURI(t *testing.T) {
	sampleM3U := `#EXTM3U
#EXTINF:-1 tvg-id="test1", Incomplete Channel`

	decoder := NewDecoder(strings.NewReader(sampleM3U))
	_, err := decoder.Decode()
	assert.Equal(t, io.EOF, err)
}

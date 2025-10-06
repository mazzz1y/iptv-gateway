package m3u8

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoderBasicEncoding(t *testing.T) {
	buffer := &bytes.Buffer{}
	encoder := NewEncoder(buffer, nil)

	uri, _ := url.Parse("http://example.com/stream1")
	track := &Track{
		Name:   "Test Channel",
		Length: -1.0,
		URI:    uri,
		Attrs: map[string]string{
			AttrTvgID:      "test1",
			AttrTvgName:    "Test Channel",
			AttrTvgLogo:    "http://example.com/logo.png",
			AttrGroupTitle: "News",
		},
	}

	err := encoder.Encode(track)
	require.NoError(t, err)

	_ = encoder.Close()

	expected := "#EXTM3U\n#EXTINF:-1 group-title=\"News\" tvg-id=\"test1\" tvg-logo=\"http://example.com/logo.png\" tvg-name=\"Test Channel\",Test Channel\nhttp://example.com/stream1\n"
	assert.Equal(t, expected, buffer.String())
}

func TestEncoderWithHeaderAttrs(t *testing.T) {
	buffer := &bytes.Buffer{}
	headerAttrs := map[string]string{
		"x-tvg-url": "http://example.com/epg.xml",
	}
	encoder := NewEncoder(buffer, headerAttrs)

	uri, _ := url.Parse("http://example.com/stream1")
	track := &Track{
		Name:   "Test Channel",
		Length: -1.0,
		URI:    uri,
	}

	err := encoder.Encode(track)
	require.NoError(t, err)

	_ = encoder.Close()

	expected := "#EXTM3U x-tvg-url=\"http://example.com/epg.xml\"\n#EXTINF:-1,Test Channel\nhttp://example.com/stream1\n"
	assert.Equal(t, expected, buffer.String())
}

func TestEncoderWithTags(t *testing.T) {
	buffer := &bytes.Buffer{}
	encoder := NewEncoder(buffer, nil)

	uri, _ := url.Parse("http://example.com/stream1")
	track := &Track{
		Name:   "Test Channel",
		Length: -1.0,
		URI:    uri,
		Attrs: map[string]string{
			AttrTvgID: "test1",
		},
		Tags: map[string]string{
			"EXTVLCOPT": "http-user-agent=Mozilla/5.0",
			"KODIPROP":  "inputstream=inputstream.adaptive",
		},
	}

	err := encoder.Encode(track)
	require.NoError(t, err)

	_ = encoder.Close()

	output := buffer.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	require.Equal(t, 5, len(lines))
	assert.Equal(t, "#EXTM3U", lines[0])
	assert.Equal(t, "#EXTINF:-1 tvg-id=\"test1\",Test Channel", lines[1])
	tagLines := []string{lines[2], lines[3]}
	assert.Contains(t, tagLines, "#EXTVLCOPT:http-user-agent=Mozilla/5.0")
	assert.Contains(t, tagLines, "#KODIPROP:inputstream=inputstream.adaptive")
	assert.Equal(t, "http://example.com/stream1", lines[4])
}

func TestEncoderMissingURI(t *testing.T) {
	buffer := &bytes.Buffer{}
	encoder := NewEncoder(buffer, nil)

	track := &Track{
		Name:   "Test Channel",
		Length: -1.0,
		Attrs: map[string]string{
			AttrTvgID: "test1",
		},
	}

	err := encoder.Encode(track)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "track missing URI")
}

func TestEncoderAfterClose(t *testing.T) {
	buffer := &bytes.Buffer{}
	encoder := NewEncoder(buffer, nil)

	err := encoder.Close()
	require.NoError(t, err)

	uri, _ := url.Parse("http://example.com/stream1")
	track := &Track{
		Name:   "Test Channel",
		Length: -1.0,
		URI:    uri,
	}

	err = encoder.Encode(track)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "encoder already closed")
}

func TestEncoderUnsupportedType(t *testing.T) {
	buffer := &bytes.Buffer{}
	encoder := NewEncoder(buffer, nil)

	err := encoder.Encode("not a track")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported item type")
}

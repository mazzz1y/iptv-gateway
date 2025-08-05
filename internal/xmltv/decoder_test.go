package xmltv

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDecoder(t *testing.T) {
	r := io.NopCloser(strings.NewReader("<xml></xml>"))
	decoder := NewDecoder(r)
	assert.NotNil(t, decoder)
	assert.IsType(t, &XMLDecoder{}, decoder)

	r2 := strings.NewReader("<xml></xml>")
	decoder2 := NewDecoder(r2)
	assert.NotNil(t, decoder2)
	assert.IsType(t, &XMLDecoder{}, decoder2)
}

func TestXMLDecoder_Decode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interface{}
		err      bool
	}{
		{
			name: "Empty TV",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<tv></tv>`,
			expected: []interface{}{
				TV{},
			},
			err: false,
		},
		{
			name: "TV with attributes",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<tv date="20230101" source-info-url="http://example.com" 
    source-info-name="Example" source-data-url="http://data.example.com" 
    generator-info-name="Generator" generator-info-url="http://generator.example.com"></tv>`,
			expected: []interface{}{
				TV{
					Date:              "20230101",
					SourceInfoURL:     "http://example.com",
					SourceInfoName:    "Example",
					SourceDataURL:     "http://data.example.com",
					GeneratorInfoName: "Generator",
					GeneratorInfoURL:  "http://generator.example.com",
				},
			},
			err: false,
		},
		{
			name: "Channel element",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <channel id="channel1">
    <display-name>Channel 1</display-name>
    <icon src="http://example.com/icon.png" width="100" height="100"/>
    <url>http://example.com/channel1</url>
  </channel>
</tv>`,
			expected: []interface{}{
				TV{},
				Channel{
					ID: "channel1",
					DisplayNames: []CommonElement{
						{Value: "Channel 1"},
					},
					Icons: []Icon{
						{
							Source: "http://example.com/icon.png",
							Width:  100,
							Height: 100,
						},
					},
					URLs: []string{"http://example.com/channel1"},
				},
			},
			err: false,
		},
		{
			name: "Programme element",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <programme start="20230101120000 +0000" stop="20230101130000 +0000" channel="channel1">
    <title>Test Programme</title>
    <desc>Programme description</desc>
    <category>Entertainment</category>
    <icon src="http://example.com/prog.png" width="100" height="100"/>
  </programme>
</tv>`,
			expected: []interface{}{
				TV{},
				Programme{
					Channel: "channel1",
					Start:   &Time{Time: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)},
					Stop:    &Time{Time: time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)},
					Titles: []CommonElement{
						{Value: "Test Programme"},
					},
					Descriptions: []CommonElement{
						{Value: "Programme description"},
					},
					Categories: []CommonElement{
						{Value: "Entertainment"},
					},
					Icons: []Icon{
						{
							Source: "http://example.com/prog.png",
							Width:  100,
							Height: 100,
						},
					},
				},
			},
			err: false,
		},
		{
			name: "Invalid XML",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <invalid>
</tv>`,
			expected: []interface{}{
				TV{},
			},
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			decoder := NewDecoder(r)

			for i, expected := range tt.expected {
				item, err := decoder.Decode()
				if i == len(tt.expected)-1 && tt.err {
					if tt.name == "Invalid XML" {
						if err != nil {
							assert.Error(t, err)
						} else {
							_, err = decoder.Decode()
							assert.Error(t, err)
						}
					} else {
						assert.Error(t, err)
					}
					continue
				}
				require.NoError(t, err)
				assert.IsType(t, expected, item)

				switch e := expected.(type) {
				case TV:
					actual := item.(TV)
					assert.Equal(t, e.Date, actual.Date)
					assert.Equal(t, e.SourceInfoURL, actual.SourceInfoURL)
					assert.Equal(t, e.SourceInfoName, actual.SourceInfoName)
					assert.Equal(t, e.SourceDataURL, actual.SourceDataURL)
					assert.Equal(t, e.GeneratorInfoName, actual.GeneratorInfoName)
					assert.Equal(t, e.GeneratorInfoURL, actual.GeneratorInfoURL)
				case Channel:
					actual := item.(Channel)
					assert.Equal(t, e.ID, actual.ID)
					assert.Equal(t, len(e.DisplayNames), len(actual.DisplayNames))
					assert.Equal(t, len(e.Icons), len(actual.Icons))
					assert.Equal(t, len(e.URLs), len(actual.URLs))
				case Programme:
					actual := item.(Programme)
					assert.Equal(t, e.Channel, actual.Channel)
					if e.Start != nil {
						assert.Equal(t, e.Start.Time.Format(time.RFC3339), actual.Start.Time.Format(time.RFC3339))
					}
					if e.Stop != nil {
						assert.Equal(t, e.Stop.Time.Format(time.RFC3339), actual.Stop.Time.Format(time.RFC3339))
					}
					assert.Equal(t, len(e.Titles), len(actual.Titles))
					assert.Equal(t, len(e.Descriptions), len(actual.Descriptions))
					assert.Equal(t, len(e.Categories), len(actual.Categories))
					assert.Equal(t, len(e.Icons), len(actual.Icons))
				}
			}

			if !tt.err {
				_, err := decoder.Decode()
				assert.Equal(t, io.EOF, err)
			}
		})
	}
}

func TestXMLDecoder_Close(t *testing.T) {
	r := io.NopCloser(strings.NewReader("<xml></xml>"))
	decoder := NewDecoder(r).(*XMLDecoder)
	err := decoder.Close()
	assert.NoError(t, err)
}

package xmltv

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEncoder(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	encoder := NewEncoder(buf)
	assert.NotNil(t, encoder)
	assert.IsType(t, &XMLEncoder{}, encoder)
}

func TestXMLEncoder_Encode(t *testing.T) {
	tests := []struct {
		name     string
		items    []any
		expected string
		err      bool
	}{
		{
			name: "Channel element",
			items: []any{
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
			expected: `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE tv SYSTEM "xmltv.dtd">
<tv><channel id="channel1"><display-name>Channel 1</display-name><icon src="http://example.com/icon.png" width="100" height="100"></icon><url>http://example.com/channel1</url></channel></tv>`,
			err: false,
		},
		{
			name: "Programme element",
			items: []any{
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
			expected: `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE tv SYSTEM "xmltv.dtd">
<tv><programme start="20230101120000+0000" stop="20230101130000+0000" channel="channel1"><title>Test Programme</title><desc>Programme description</desc><category>Entertainment</category><icon src="http://example.com/prog.png" width="100" height="100"></icon></programme></tv>`,
			err: false,
		},
		{
			name: "Multiple items",
			items: []any{
				Channel{
					ID: "channel1",
					DisplayNames: []CommonElement{
						{Value: "Channel 1"},
					},
				},
				Programme{
					Channel: "channel1",
					Start:   &Time{Time: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)},
					Titles:  []CommonElement{{Value: "Test Programme"}},
				},
			},
			expected: `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE tv SYSTEM "xmltv.dtd">
<tv><channel id="channel1"><display-name>Channel 1</display-name></channel><programme start="20230101120000+0000" channel="channel1"><title>Test Programme</title></programme></tv>`,
			err: false,
		},
		{
			name:  "Invalid item type",
			items: []any{"invalid"},
			expected: `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE tv SYSTEM "xmltv.dtd">
<tv></tv>`,
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			encoder := NewEncoder(buf)

			for i, item := range tt.items {
				err := encoder.Encode(item)
				if i == len(tt.items)-1 && tt.err {
					assert.Error(t, err)
					continue
				}
				require.NoError(t, err)
			}

			err := encoder.WriteFooter()
			require.NoError(t, err)

			err = encoder.Close()
			require.NoError(t, err)

			xml := buf.String()
			xml = strings.ReplaceAll(xml, " ", "")
			xml = strings.ReplaceAll(xml, "\n", "")
			xml = strings.ReplaceAll(xml, "\t", "")
			expected := strings.ReplaceAll(tt.expected, " ", "")
			expected = strings.ReplaceAll(expected, "\n", "")

			assert.Equal(t, expected, xml)
		})
	}
}

func TestXMLEncoder_WriteHeaderFooter(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	encoder := NewEncoder(buf)

	err := encoder.writeHeader()
	require.NoError(t, err)
	assert.True(t, encoder.headerWritten)

	pos := buf.Len()
	err = encoder.writeHeader()
	require.NoError(t, err)
	assert.Equal(t, pos, buf.Len(), "Header was written twice")

	err = encoder.WriteFooter()
	require.NoError(t, err) // footer already written
	assert.True(t, encoder.footerWritten)

	err = encoder.WriteFooter()
	require.Error(t, err)

	buf2 := bytes.NewBuffer(nil)
	encoder2 := NewEncoder(buf2)
	encoder2.headerWritten = false

	err = encoder2.WriteFooter()
	require.Error(t, err) // header not written
}

func TestXMLEncoder_Close(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	encoder := NewEncoder(buf)

	err := encoder.Encode(Channel{ID: "test"})
	require.NoError(t, err)

	err = encoder.WriteFooter()
	require.NoError(t, err)

	err = encoder.Close()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "</tv>")
}

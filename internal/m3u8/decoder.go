package m3u8

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	extinfRegex            = regexp.MustCompile(`#EXTINF:(-?\d+(\.\d+)?)\s*(.*)$`)
	attrRegex              = regexp.MustCompile(`([a-zA-Z0-9_-]+)="([^"]*)"`)
	whitelistedTagPrefixes = []string{
		"#EXT-X-",
		"#EXTGRP:",
		"#EXTBYT:",
		"#EXTSIZE:",
		"#EXTBIN:",
		"#EXTVLCOPT:",
	}
)

type Decoder interface {
	Decode() (any, error)
	Close() error
}

type M3UDecoder struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
	header  bool
	done    bool
}

func NewDecoder(r io.Reader) Decoder {
	var readCloser io.ReadCloser
	if rc, ok := r.(io.ReadCloser); ok {
		readCloser = rc
	} else {
		readCloser = io.NopCloser(r)
	}

	return &M3UDecoder{
		reader:  readCloser,
		scanner: bufio.NewScanner(r),
		header:  false,
		done:    false,
	}
}

func (d *M3UDecoder) Decode() (any, error) {
	if d.done {
		return nil, io.EOF
	}

	if !d.header {
		if !d.scanner.Scan() {
			if err := d.scanner.Err(); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}

		line := strings.TrimSpace(d.scanner.Text())
		if !strings.HasPrefix(line, "#EXTM3U") {
			return nil, fmt.Errorf("invalid M3U file format, missing #EXTM3U header")
		}

		d.header = true
	}

	track, err := d.parseNextTrack()
	if err != nil {
		if err == io.EOF {
			d.done = true
		}
		return nil, err
	}

	return track, nil
}

func (d *M3UDecoder) parseNextTrack() (*Track, error) {
	var track *Track

	for d.scanner.Scan() {
		line := strings.TrimSpace(d.scanner.Text())

		if line == "" || (strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#EXT")) {
			continue
		}

		if strings.HasPrefix(line, "#EXTINF:") {
			var err error
			track, err = parseExtInfLine(line)
			if err != nil {
				return nil, fmt.Errorf("error parsing EXTINF line: %w", err)
			}
			continue
		}

		if track != nil && hasWhitelistedTagPrefix(line) {
			tagMap := parseTags(line)
			for key, value := range tagMap {
				track.Tags[key] = value
			}
			continue
		}

		if track != nil && !strings.HasPrefix(line, "#") {
			u, err := url.Parse(line)
			if err != nil {
				return nil, fmt.Errorf("invalid URL: %w", err)
			}
			track.URI = u
			return track, nil
		}
	}

	if err := d.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

func (d *M3UDecoder) Close() error {
	return d.reader.Close()
}

func parseExtInfLine(line string) (*Track, error) {
	matches := extinfRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid EXTINF format")
	}

	duration, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil, err
	}

	track := &Track{
		Length: duration,
		Attrs:  make(map[string]string),
		Tags:   make(map[string]string),
	}

	remainingContent := matches[3]

	attrMatches := attrRegex.FindAllStringSubmatch(remainingContent, -1)
	for _, match := range attrMatches {
		if len(match) > 2 {
			track.Attrs[match[1]] = match[2]
		}
	}

	titleParts := strings.Split(remainingContent, ",")
	if len(titleParts) > 1 {
		track.Name = strings.TrimSpace(titleParts[len(titleParts)-1])
	} else if len(titleParts) == 1 && len(attrMatches) == 0 {
		track.Name = strings.TrimSpace(remainingContent)
	}

	return track, nil
}

func hasWhitelistedTagPrefix(line string) bool {
	for _, prefix := range whitelistedTagPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func parseTags(line string) map[string]string {
	t := make(map[string]string)

	for _, prefix := range whitelistedTagPrefixes {
		if !strings.HasPrefix(line, prefix) {
			continue
		}

		var key, value string
		if prefix == "#EXT-X-" {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				key = strings.TrimPrefix(parts[0], "#")
				value = parts[1]
			}
		} else {
			key = strings.TrimSuffix(strings.TrimPrefix(prefix, "#"), ":")
			value = strings.TrimPrefix(line, prefix)
		}

		if key != "" {
			t[key] = value
			return t
		}
	}

	return t
}

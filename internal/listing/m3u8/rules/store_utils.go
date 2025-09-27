package rules

import (
	"bytes"
	"regexp"
	"strings"

	"iptv-gateway/internal/config/types"
)

func extractBaseName(fv string, patterns []*regexp.Regexp) string {
	if fv == "" {
		return ""
	}
	if len(patterns) == 0 {
		return fv
	}

	result := fv
	for _, pattern := range patterns {
		result = pattern.ReplaceAllString(result, "")
	}

	return strings.Join(strings.Fields(result), " ")
}

func getPatterns(
	namePatterns types.RegexpArr, attrPatterns *types.NamePatterns, tagPatterns *types.NamePatterns) []*regexp.Regexp {
	if len(namePatterns) > 0 {
		return namePatterns.ToArray()
	}
	if attrPatterns != nil {
		return attrPatterns.Patterns.ToArray()
	}
	if tagPatterns != nil {
		return tagPatterns.Patterns.ToArray()
	}
	return nil
}

func setFieldValue(
	ch *Channel, value string, namePatterns types.RegexpArr, attrPatterns *types.NamePatterns, tagPatterns *types.NamePatterns) {
	if len(namePatterns) > 0 {
		ch.SetName(value)
	} else if attrPatterns != nil {
		ch.SetAttr(attrPatterns.Name, value)
	} else if tagPatterns != nil {
		ch.SetTag(tagPatterns.Name, value)
	} else {
		ch.SetName(value)
	}
}

func getFieldValue(
	ch *Channel, namePatterns types.RegexpArr, attrPatterns *types.NamePatterns, tagPatterns *types.NamePatterns) string {
	if len(namePatterns) > 0 {
		return ch.Name()
	}
	if attrPatterns != nil {
		if val, ok := ch.GetAttr(attrPatterns.Name); ok {
			return val
		}
	}
	if tagPatterns != nil {
		if val, ok := ch.GetTag(tagPatterns.Name); ok {
			return val
		}
	}
	return ch.Name()
}

func hasPatternVariationsGroup(
	group []*Channel, namePatterns types.RegexpArr, attrPatterns *types.NamePatterns, tagPatterns *types.NamePatterns) bool {
	patterns := getPatterns(namePatterns, attrPatterns, tagPatterns)
	if len(patterns) == 0 {
		return false
	}

	hasMultiplePatterns := false

	for _, ch := range group {
		fv := getFieldValue(ch, namePatterns, attrPatterns, tagPatterns)
		for _, pattern := range patterns {
			patternStr := pattern.String()
			if patternStr == "" {
				continue
			}
			if pattern.MatchString(fv) && patternStr != fv {
				hasMultiplePatterns = true
				break
			}
		}
		if hasMultiplePatterns {
			break
		}
	}

	return hasMultiplePatterns
}

func selectBestChannel(channels []*Channel, namePatterns types.RegexpArr, attrPatterns *types.NamePatterns, tagPatterns *types.NamePatterns) *Channel {
	patterns := getPatterns(namePatterns, attrPatterns, tagPatterns)
	if len(patterns) == 0 {
		return channels[0]
	}

	for _, pattern := range patterns {
		for _, ch := range channels {
			fv := getFieldValue(ch, namePatterns, attrPatterns, tagPatterns)
			if pattern.MatchString(fv) {
				return ch
			}
		}
	}
	return channels[0]
}

func processSetField(ch *Channel, setField *types.Template, baseName string) string {
	if setField == nil {
		return baseName
	}

	tmplMap := map[string]any{
		"Channel": map[string]any{
			"Name":  ch.Name(),
			"Attrs": ch.Attrs(),
			"Tags":  ch.Tags(),
		},
		"BaseName": baseName,
	}

	var buf bytes.Buffer
	if err := setField.ToTemplate().Execute(&buf, tmplMap); err != nil {
		return baseName
	}

	return buf.String()
}

func evaluateStoreWhenCondition(when *types.Condition, clientName string) bool {
	if when == nil {
		return true
	}

	if len(when.Clients) > 0 {
		for _, client := range when.Clients {
			if clientName == client {
				return !when.Invert
			}
		}
		return when.Invert
	}

	return true
}

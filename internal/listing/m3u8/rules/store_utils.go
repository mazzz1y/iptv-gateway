package rules

import (
	"regexp"
	"strings"

	"iptv-gateway/internal/config/common"
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

func extractBaseNameFromChannel(ch *Channel, selector *common.Selector, patterns common.RegexpArr) (string, bool) {
	fv, ok := getSelectorFieldValue(ch, selector)
	if !ok {
		return "", false
	}
	patternArray := patterns.ToArray()
	return extractBaseName(fv, patternArray), true
}

func getSelectorFieldValue(ch *Channel, selector *common.Selector) (string, bool) {
	if selector == nil {
		return ch.Name(), true
	}

	switch selector.Type {
	case common.SelectorName:
		return ch.Name(), true
	case common.SelectorAttr:
		if val, ok := ch.GetAttr(selector.Value); ok {
			return val, true
		}
		return "", false
	case common.SelectorTag:
		if val, ok := ch.GetTag(selector.Value); ok {
			return val, true
		}
		return "", false
	}

	return "", false
}

func hasPatternVariationsGroup(group []*Channel, selector *common.Selector, patterns common.RegexpArr) bool {
	patternArray := patterns.ToArray()
	if len(patternArray) == 0 {
		return false
	}

	hasMultiplePatterns := false

	for _, ch := range group {
		fv, ok := getSelectorFieldValue(ch, selector)
		if !ok || fv == "" {
			continue
		}
		for _, pattern := range patternArray {
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

func selectBestChannel(channels []*Channel, selector *common.Selector, patterns common.RegexpArr) *Channel {
	patternArray := patterns.ToArray()
	if len(patternArray) == 0 {
		return channels[0]
	}

	for _, pattern := range patternArray {
		for _, ch := range channels {
			fv, ok := getSelectorFieldValue(ch, selector)
			if ok && pattern.MatchString(fv) {
				return ch
			}
		}
	}
	return channels[0]
}

func evaluateStoreCondition(condition *common.Condition, clientName string) bool {
	if condition == nil {
		return true
	}

	if len(condition.Clients) > 0 {
		for _, client := range condition.Clients {
			if clientName == client {
				return !condition.Invert
			}
		}
		return condition.Invert
	}

	return true
}

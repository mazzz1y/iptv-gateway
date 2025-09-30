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

func getSelectorFieldValue(ch *Channel, selector *common.Selector) string {
	if selector == nil {
		return ch.Name()
	}

	switch selector.Type {
	case common.SelectorName:
		return ch.Name()
	case common.SelectorAttr:
		if val, ok := ch.GetAttr(selector.Value); ok {
			return val
		}
	case common.SelectorTag:
		if val, ok := ch.GetTag(selector.Value); ok {
			return val
		}
	}
	return ch.Name()
}

func hasPatternVariationsGroup(group []*Channel, selector *common.Selector, patterns common.RegexpArr) bool {
	patternArray := patterns.ToArray()
	if len(patternArray) == 0 {
		return false
	}

	hasMultiplePatterns := false

	for _, ch := range group {
		fv := getSelectorFieldValue(ch, selector)
		if fv == "" {
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
			fv := getSelectorFieldValue(ch, selector)
			if pattern.MatchString(fv) {
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

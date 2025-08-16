package actions

import (
	"iptv-gateway/internal/listing/m3u8/channel"
	"regexp"
	"strings"
)

type RemoveDuplicatesRule struct {
	patterns    []string
	regexps     []*regexp.Regexp
	trimPattern bool
}

func NewRemoveDuplicatesRule(patterns []string, trimPattern bool) *RemoveDuplicatesRule {
	regexps := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		regexps[i] = regexp.MustCompile(pattern)
	}

	return &RemoveDuplicatesRule{
		patterns:    patterns,
		regexps:     regexps,
		trimPattern: trimPattern,
	}
}

func (r *RemoveDuplicatesRule) Apply(global *channel.Registry, sub *channel.Registry) {
	globalGroupedByBaseName := make(map[string][]*channel.Channel)

	for _, ch := range global.All() {
		baseName := r.extractBaseName(ch.Name())
		globalGroupedByBaseName[baseName] = append(globalGroupedByBaseName[baseName], ch)
	}

	r.processDuplicateGroups(globalGroupedByBaseName, sub)
}

func (r *RemoveDuplicatesRule) extractBaseName(name string) string {
	for _, regex := range r.regexps {
		name = regex.ReplaceAllString(name, "")
	}
	return strings.TrimSpace(name)
}

func (r *RemoveDuplicatesRule) selectBestChannel(channels []*channel.Channel) *channel.Channel {
	for _, regex := range r.regexps {
		for _, ch := range channels {
			if regex.MatchString(ch.Name()) {
				return ch
			}
		}
	}
	return channels[0]
}

func (r *RemoveDuplicatesRule) processDuplicateGroups(globalGroupedByBaseName map[string][]*channel.Channel, subscriptionStore *channel.Registry) {
	subscriptionChannels := make(map[*channel.Channel]bool)
	for _, ch := range subscriptionStore.All() {
		subscriptionChannels[ch] = true
	}

	for baseName, globalChannels := range globalGroupedByBaseName {
		if len(globalChannels) <= 1 {
			continue
		}

		bestChannel := r.selectBestChannel(globalChannels)

		subscriptionChannelsInGroup := make([]*channel.Channel, 0)
		for _, ch := range globalChannels {
			if subscriptionChannels[ch] {
				subscriptionChannelsInGroup = append(subscriptionChannelsInGroup, ch)
			}
		}

		if len(subscriptionChannelsInGroup) > 0 {
			bestFromSubscription := false
			for _, ch := range subscriptionChannelsInGroup {
				if ch == bestChannel {
					bestFromSubscription = true
					if r.trimPattern {
						ch.Track().Name = baseName
					}
					break
				}
			}

			for _, ch := range subscriptionChannelsInGroup {
				if !bestFromSubscription || ch != bestChannel {
					ch.Track().IsRemoved = true
				}
			}
		}
	}
}

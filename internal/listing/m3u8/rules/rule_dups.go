package rules

import (
	"regexp"
	"strings"
)

type RemoveDuplicatesRule struct {
	patterns    []*regexp.Regexp
	trimPattern bool
}

func NewRemoveDuplicatesRule(patterns []*regexp.Regexp, trimPattern bool) *RemoveDuplicatesRule {
	return &RemoveDuplicatesRule{
		patterns:    patterns,
		trimPattern: trimPattern,
	}
}

func (r *RemoveDuplicatesRule) Apply(global *Store, sub *Store) {
	globalGroupedByBaseName := make(map[string][]*Channel)

	for _, ch := range global.All() {
		baseName := r.extractBaseName(ch.Name())
		globalGroupedByBaseName[baseName] = append(globalGroupedByBaseName[baseName], ch)
	}

	r.processDuplicateGroups(globalGroupedByBaseName, sub)
}

func (r *RemoveDuplicatesRule) extractBaseName(name string) string {
	for _, regex := range r.patterns {
		name = regex.ReplaceAllString(name, "")
	}

	return strings.Join(strings.Fields(name), " ")
}

func (r *RemoveDuplicatesRule) selectBestChannel(channels []*Channel) *Channel {
	for _, regex := range r.patterns {
		for _, ch := range channels {
			if regex.MatchString(ch.Name()) {
				return ch
			}
		}
	}

	return channels[0]
}

func (r *RemoveDuplicatesRule) processDuplicateGroups(globalGroupedByBaseName map[string][]*Channel, subscriptionStore *Store) {
	subscriptionChannels := make(map[*Channel]bool)
	for _, ch := range subscriptionStore.All() {
		subscriptionChannels[ch] = true
	}

	for baseName, globalChannels := range globalGroupedByBaseName {
		if len(globalChannels) <= 1 {
			continue
		}

		bestChannel := r.selectBestChannel(globalChannels)

		subscriptionChannelsInGroup := make([]*Channel, 0)
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

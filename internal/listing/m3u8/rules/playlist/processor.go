package playlist

import (
	"context"
	"majmun/internal/config/common"
	playlistconf "majmun/internal/config/rules/playlist"
	"majmun/internal/listing/m3u8/store"
)

type Processor struct {
	rules      []*playlistconf.Rule
	clientName string
}

func NewRulesProcessor(clientName string, rules []*playlistconf.Rule) *Processor {
	return &Processor{
		rules:      rules,
		clientName: clientName,
	}
}

func (p *Processor) Apply(ctx context.Context, store *store.Store) {
	for _, rule := range p.rules {
		if ctx.Err() != nil {
			return
		}
		if rule.MergeChannels != nil && evaluateStoreCondition(rule.MergeChannels.Condition, p.clientName) {
			processor := NewMergeDuplicatesActionProcessor(rule.MergeChannels)
			processor.Apply(store)
		}
		if rule.RemoveDuplicates != nil && evaluateStoreCondition(rule.RemoveDuplicates.Condition, p.clientName) {
			processor := NewRemoveDuplicatesActionProcessor(rule.RemoveDuplicates)
			processor.Apply(store)
		}
		if rule.SortRule != nil && evaluateStoreCondition(rule.SortRule.Condition, p.clientName) {
			processor := NewSortProcessor(rule.SortRule)
			processor.Apply(store)
		}
	}
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

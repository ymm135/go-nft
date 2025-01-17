/*
 * This file is part of the go-nft project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2021 Red Hat, Inc.
 *
 */

package testlib

import (
	"sort"
	"testing"

	"github.com/ymm135/go-nft/nft"
	"github.com/ymm135/go-nft/nft/schema"
)

func RunTestWithFlushTable(t *testing.T, test func(t *testing.T)) {
	test(t)

	nft.ApplyConfig(&nft.Config{schema.Root{Nftables: []schema.Nftable{
		{Flush: &schema.Objects{Ruleset: true}},
	}}})
}

// NormalizeConfigForComparison returns the configuration ready for comparison with another by
// - removing the metainfo entry.
// - removing the handle + index parameters.
// - Sorting the list.
func NormalizeConfigForComparison(config *nft.Config) *nft.Config {
	if len(config.Nftables) > 0 && config.Nftables[0].Metainfo != nil {
		config.Nftables = config.Nftables[1:]
	}

	for _, nftable := range config.Nftables {
		if nftable.Rule != nil {
			nftable.Rule.Index = nil
			nftable.Rule.Handle = nil
		}
	}

	sort.Slice(config.Nftables, func(i int, j int) bool {
		s := config.Nftables
		isTableFirst := s[i].Table != nil && (s[j].Chain != nil || s[j].Rule != nil)
		isChainBeforeRule := s[i].Chain != nil && s[j].Rule != nil
		return isTableFirst || isChainBeforeRule
	})
	return config
}

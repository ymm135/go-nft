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

package nft_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eddev/go-nft/nft"
	"github.com/eddev/go-nft/nft/schema"
)

type ruleAction string

// Rule Actions
const (
	ruleADD    ruleAction = "add"
	ruleDELETE ruleAction = "delete"
)

func TestRule(t *testing.T) {
	testAddRuleWithMatchAndVerdict(t)
	testDeleteRule(t)
}

func testAddRuleWithMatchAndVerdict(t *testing.T) {
	const comment = "mycomment"

	table := nft.NewTable(tableName, nft.FamilyIP)
	chain := nft.NewRegularChain(table, chainName)

	t.Run("Add rule with match and verdict, check serialization", func(t *testing.T) {
		statements, serializedStatements := matchSrcIP4withReturnVerdict()
		rule := nft.NewRule(table, chain, statements, nil, comment)

		config := nft.NewConfig()
		config.AddRule(rule)

		serializedConfig, err := config.MarshalJSON()
		assert.NoError(t, err)

		expectedConfig := buildSerializedConfig(ruleADD, serializedStatements, nil, comment)
		assert.Equal(t, string(expectedConfig), string(serializedConfig))
	})

	t.Run("Add rule with match and verdict, check deserialization", func(t *testing.T) {
		statements, serializedStatements := matchSrcIP4withReturnVerdict()

		serializedConfig := buildSerializedConfig(ruleADD, serializedStatements, nil, comment)

		var deserializedConfig nft.Config
		assert.NoError(t, json.Unmarshal(serializedConfig, &deserializedConfig))

		rule := nft.NewRule(table, chain, statements, nil, comment)
		expectedConfig := nft.NewConfig()
		expectedConfig.AddRule(rule)

		assert.Equal(t, expectedConfig, &deserializedConfig)
	})
}

func testDeleteRule(t *testing.T) {
	table := nft.NewTable(tableName, nft.FamilyIP)
	chain := nft.NewRegularChain(table, chainName)

	t.Run("Delete rule", func(t *testing.T) {
		handleID := 100
		rule := nft.NewRule(table, chain, nil, &handleID, "")

		config := nft.NewConfig()
		config.DeleteRule(rule)

		serializedConfig, err := config.MarshalJSON()
		assert.NoError(t, err)

		expectedConfig := buildSerializedConfig(ruleDELETE, "", &handleID, "")
		assert.Equal(t, string(expectedConfig), string(serializedConfig))
	})
}

func buildSerializedConfig(action ruleAction, serializedStatements string, handle *int, comment string) []byte {
	ruleArgs := fmt.Sprintf(`"family":%q,"table":%q,"chain":%q`, nft.FamilyIP, tableName, chainName)
	if serializedStatements != "" {
		ruleArgs += "," + serializedStatements
	}
	if handle != nil {
		ruleArgs += fmt.Sprintf(`,"handle":%d`, *handle)
	}
	if comment != "" {
		ruleArgs += fmt.Sprintf(`,"comment":%q`, comment)
	}

	var config string
	if action == ruleADD {
		config = fmt.Sprintf(`{"nftables":[{"rule":{%s}}]}`, ruleArgs)
	} else {
		config = fmt.Sprintf(`{"nftables":[{%q:{"rule":{%s}}}]}`, action, ruleArgs)
	}
	return []byte(config)
}

func matchSrcIP4withReturnVerdict() ([]schema.Statement, string) {
	ipAddress := "10.10.10.10"
	matchSrcIP4 := schema.Statement{
		Match: &schema.Match{
			Op: schema.OperEQ,
			Left: schema.Expression{
				Payload: &schema.Payload{
					Protocol: schema.PayloadProtocolIP4,
					Field:    schema.PayloadFieldIPSAddr,
				},
			},
			Right: schema.Expression{String: &ipAddress},
		},
	}

	verdict := schema.Statement{}
	verdict.Return = true

	statements := []schema.Statement{matchSrcIP4, verdict}

	expectedMatch := fmt.Sprintf(
		`"match":{"op":"==","left":{"payload":{"protocol":"ip","field":"saddr"}},"right":%q}`, ipAddress,
	)
	expectedVerdict := `"return":null`
	serializedStatements := fmt.Sprintf(`"expr":[{%s},{%s}]`, expectedMatch, expectedVerdict)

	return statements, serializedStatements
}

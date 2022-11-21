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

package tests

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/ymm135/go-nft/tests/testlib"

	"github.com/ymm135/go-nft/nft"
	"github.com/ymm135/go-nft/nft/schema"
)

func TestConfig(t *testing.T) {
	testlib.RunTestWithFlushTable(t, testReadEmptyConfig)
	testlib.RunTestWithFlushTable(t, testApplyConfigWithAnEmptyTable)
	testlib.RunTestWithFlushTable(t, testApplyConfigWithSampleStatements)
}

func testReadEmptyConfig(t *testing.T) {
	config, err := nft.ReadConfig()
	assert.NoError(t, err)
	assert.Len(t, config.Nftables, 1, "Expecting just the metainfo entry")

	expectedConfig := nft.NewConfig()
	expectedConfig.Nftables = append(expectedConfig.Nftables, schema.Nftable{Metainfo: &schema.Metainfo{JsonSchemaVersion: 1}})

	// The underling nftable userspace version depends on where it is ran, therefore remove it from the diff.
	expectedConfig.Nftables[0].Metainfo.Version = config.Nftables[0].Metainfo.Version
	expectedConfig.Nftables[0].Metainfo.ReleaseName = config.Nftables[0].Metainfo.ReleaseName
	assert.Equal(t, expectedConfig, config)
}

func testApplyConfigWithAnEmptyTable(t *testing.T) {
	config := nft.NewConfig()
	config.AddTable(nft.NewTable("mytable", nft.FamilyIP))

	assert.NoError(t, nft.ApplyConfig(config))

	newConfig, err := nft.ReadConfig()
	assert.NoError(t, err)

	assert.Len(t, newConfig.Nftables, 2, "Expecting the metainfo and an empty table entry")
	assert.Equal(t, config.Nftables[0], newConfig.Nftables[1])
}

func testApplyConfigWithSampleStatements(t *testing.T) {
	testApplyConfigWithStatements(t,
		schema.Statement{Counter: &schema.Counter{}},
	)
}

func testApplyConfigWithStatements(t *testing.T, statements ...schema.Statement) {
	const tableName = "mytable"
	config := nft.NewConfig()
	table := nft.NewTable(tableName, nft.FamilyIP)
	config.AddTable(table)

	const chainName = "mychain"
	chain := nft.NewChain(table, chainName, nil, nil, nil, nil)
	config.AddChain(chain)

	rule := nft.NewRule(table, chain, statements, nil, nil, "test")
	config.AddRule(rule)

	assert.NoError(t, nft.ApplyConfig(config))

	newConfig, err := nft.ReadConfig()
	assert.NoError(t, err)

	config = testlib.NormalizeConfigForComparison(config)
	newConfig = testlib.NormalizeConfigForComparison(newConfig)
	assert.Equal(t, config.Nftables, newConfig.Nftables)
}

func TestApplyConfigWithRawJSON(t *testing.T) {
	const tableName = "my-table"
	config := nft.NewConfig()
	table := nft.NewTable(tableName, nft.FamilyIP)
	config.AddTable(table)

	const chainName = "mychain"
	chain := nft.NewChain(table, chainName, nil, nil, nil, nil)
	config.AddChain(chain)

	ifaceName := "nic0"
	// {Verdict: schema.Accept()},
	statements := []schema.Statement{
		{Match: &schema.Match{
			Op:    schema.OperEQ,
			Left:  schema.Expression{RowData: []byte(`{"meta":{"key":"iifname"}}`)},
			Right: schema.Expression{String: &ifaceName},
		}},
		{Match: &schema.Match{
			Op:   schema.OperEQ,
			Left: schema.Expression{Payload: &schema.Payload{Protocol: schema.PayloadProtocolIP4, Field: schema.PayloadFieldIPSAddr}},
			Right: schema.Expression{Set: &[]schema.SetElement{
				{Value: "192.168.0.1"},
				{Value: "192.168.0.2"},
				{Range: &schema.Range{Range: [2]string{"192.168.1.0", "192.168.1.100"}}},
				{IPPrefix: &schema.IPPrefix{Prefix: schema.Prefix{Addr: "192.168.2.1", Len: 24}}}}},
		}},
		{Match: &schema.Match{
			Op:   schema.OperEQ,
			Left: schema.Expression{RowData: []byte(`{"meta":{"key":"time"}}`)},
			Right: schema.Expression{Set: &[]schema.SetElement{
				{Range: &schema.Range{Range: [2]string{"2022-11-01 18:00:00", "2022-11-01 19:00:00"}}},
				{Range: &schema.Range{Range: [2]string{"2022-11-10 18:00:00", "2022-11-10 19:00:00"}}},
			}},
		}},
		{RawData: []byte(`{"log": { "prifix":"accept-log"} }`)},
		{RawData: []byte(`{"accept": null}`)},
	}

	rule := nft.NewRule(table, chain, statements, nil, nil, "test")
	config.AddRule(rule)

	assert.NoError(t, nft.ApplyConfig(config))

	//newConfig, err := nft.ReadConfig()
	//assert.NoError(t, err)
	//
	//config = testlib.NormalizeConfigForComparison(config)
	//newConfig = testlib.NormalizeConfigForComparison(newConfig)
	//assert.Equal(t, config.Nftables, newConfig.Nftables)
}

package xmlparsing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type parseTestCase struct {
	name         string
	rawXml       string
	expectedJson string
}

func (tc *parseTestCase) runTest(t *testing.T) {
	data, err := Parse(tc.rawXml)
	if assert.NoError(t, err) {
		j, _ := json.Marshal(data)
		assert.JSONEq(t, tc.expectedJson, string(j))
	}
}

var parseTestCases = []parseTestCase{
	{
		name: "TestCase#1",
		rawXml: `
		<?xml version="1.0" encoding="UTF-8"?>
		<Person>
			<Company><![CDATA[Hogwarts School of Witchcraft and Wizardry]]></Company>
			<FullName>Harry Potter</FullName>
			<Email where="work">hpotter@hog.gb</Email>
			<Email where="home">hpotter@gmail.com</Email>
			<Addr>4 Privet Drive</Addr>
			<Group>
				<Value>Hexes</Value>
				<Value>Jinxes</Value>
			</Group>
		</Person>
		`,
		expectedJson: `
		{
			"Person": {
				"Company": "Hogwarts School of Witchcraft and Wizardry",
				"FullName": "Harry Potter",
				"Email": [
				{
					"-attrs": {"where": "work"},
					"content": "hpotter@hog.gb"
				},
				{
					"-attrs": {"where": "home"},
					"content": "hpotter@gmail.com"
				}
				],
				"Addr": "4 Privet Drive",
				"Group": {
					"Value": ["Hexes", "Jinxes"]
				}
			}
		}
		`,
	},
	{
		name: "TestCase#2_namespaces",
		rawXml: `
		<ns1:person>
			<ns2:name>Eddie</ns2:name>
			<ns2:surname>Dean</ns2:surname>
		</ns1:person>
		`,
		expectedJson: `
		{
			"ns1:person": {
				"ns2:name": "Eddie",
				"ns2:surname": "Dean"
			}
		}
		`,
	},
	{
		name:   "TestCase#3_emptytag",
		rawXml: "<body><emptytag/></body>",
		expectedJson: `{
			"body": {
				"emptytag": ""
			}
		}
		`,
	},
	{
		name:   "TestCase#4_onlyattributes",
		rawXml: `<body><tag attr1="attr1_value" attr2="attr2_value"/></body>`,
		expectedJson: `{
			"body": {
				"tag": {
					"-attrs": {
						"attr1": "attr1_value",
						"attr2": "attr2_value"
					},
					"content": ""
				}
			}
		}
		`,
	},
}

func TestParse(t *testing.T) {
	for _, tc := range parseTestCases {
		t.Run(tc.name, tc.runTest)
	}
}

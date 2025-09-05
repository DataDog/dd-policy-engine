package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestParserValidInputs(t *testing.T) {
	var tests = []struct {
		name         string
		input        string
		expectedJSON string
	}{
		{
			name:         "simple term 1/x",
			input:        "foo:bar",
			expectedJSON: `{"Field":"foo","Value":"bar","Comparator":"EXACT"}`,
		},
		{
			name:         "simple term 2/x",
			input:        "    foo:bar",
			expectedJSON: `{"Field":"foo","Value":"bar","Comparator":"EXACT"}`,
		},
		{
			name:         "simple term 3/x",
			input:        "foo:bar     ",
			expectedJSON: `{"Field":"foo","Value":"bar","Comparator":"EXACT"}`,
		},
		{
			name:         "simple term 4/x",
			input:        "  foo:bar     ",
			expectedJSON: `{"Field":"foo","Value":"bar","Comparator":"EXACT"}`,
		},
		// // {"simple term: prefix matching", "foo:*b"},
		{
			name:         "implicit and terms 1/x",
			input:        "hello:world second:term",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"hello","Value":"world","Comparator":"EXACT"},"Right":{"Field":"second","Value":"term","Comparator":"EXACT"},"Comparator":"AND"}}`,
		},
		{
			name:         "implicit and terms 2/x",
			input:        "hello:world         second:term",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"hello","Value":"world","Comparator":"EXACT"},"Right":{"Field":"second","Value":"term","Comparator":"EXACT"},"Comparator":"AND"}}`,
		},
		{
			name:         "implicit and terms 3/x",
			input:        "hello:world andsecond:term",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"hello","Value":"world","Comparator":"EXACT"},"Right":{"Field":"andsecond","Value":"term","Comparator":"EXACT"},"Comparator":"AND"}}`,
		},
		{
			name:         "explicit and terms 1/x",
			input:        "team:apm AND lang:fr",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"team","Value":"apm","Comparator":"EXACT"},"Right":{"Field":"lang","Value":"fr","Comparator":"EXACT"},"Comparator":"AND"}}`,
		},
		{
			name:         "explicit and terms 2/x",
			input:        "team:apm    AND lang:fr",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"team","Value":"apm","Comparator":"EXACT"},"Right":{"Field":"lang","Value":"fr","Comparator":"EXACT"},"Comparator":"AND"}}`,
		},
		{
			name:         "explicit and terms 3/x",
			input:        "team:apm AND   lang:fr",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"team","Value":"apm","Comparator":"EXACT"},"Right":{"Field":"lang","Value":"fr","Comparator":"EXACT"},"Comparator":"AND"}}`,
		},
		{
			name:         "explicit and terms 4/x",
			input:        "team:apm      AND   lang:fr",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"team","Value":"apm","Comparator":"EXACT"},"Right":{"Field":"lang","Value":"fr","Comparator":"EXACT"},"Comparator":"AND"}}`,
		},
		{
			name:         "explicit and terms 5/x",
			input:        "team:apm and lang:fr",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"team","Value":"apm","Comparator":"EXACT"},"Right":{"Field":"lang","Value":"fr","Comparator":"EXACT"},"Comparator":"AND"}}`,
		},
		{
			name:         "explicit or terms 1/x",
			input:        "k:v or k1:v2",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"k","Value":"v","Comparator":"EXACT"},"Right":{"Field":"k1","Value":"v2","Comparator":"EXACT"},"Comparator":"OR"}}`,
		},
		{
			name:         "explicit or terms 2/x",
			input:        "k:v ork1:v2",
			expectedJSON: `{"BooleanNode":{"Left":{"Field":"k","Value":"v","Comparator":"EXACT"},"Right":{"Field":"ork1","Value":"v2","Comparator":"EXACT"},"Comparator":"AND"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			if err != nil {
				t.Fatalf(fmt.Sprintf("unexpected parsing error:\n %s", err))
			}

			var obj1, obj2 any
			json.Unmarshal([]byte(tt.expectedJSON), &obj1)

			json_ast, err := json.Marshal(ast)
			json.Unmarshal([]byte(json_ast), &obj2)

			if !reflect.DeepEqual(obj1, obj2) {
				t.Fatalf(fmt.Sprintf("AST differs: \nexpected: %s\n  actual: %s", tt.expectedJSON, json_ast))
			}
		})
	}
}

// Support unlimited spaces between AND/OR/implicit and

func TestParserInvalidInputs(t *testing.T) {
	var tests = []struct {
		name  string
		input string
	}{
		{name: "missing key", input: ":value"},
		{name: "missing \":\" separator", input: "keywithoutvalue"},
		{name: "missing value 1/x", input: "key:"},
		{name: "missing value 2/x", input: "key:  "},
		{name: "key must start with a letter 1/x", input: "1key:value"},
		{name: "key must start with a letter 2/x", input: "key:value  8"},
		{name: "key doesn't accept spaces", input: "space in key:value"},
		{name: "edge case 1/x", input: "k:v and"},
		{name: "edge case 2/x", input: "k:v or"},
		{name: "edge case 3/x", input: "k:v and k2:"},
		{name: "edge case 4/x", input: "k:v and :v2"},
		{name: "edge case 5/x", input: "k:v and :"},
		{name: "edge case 6/x", input: "k:v or k2:"},
		{name: "edge case 7/x", input: "k:v or :v2"},
		{name: "edge case 8/x", input: "k:v or :"},
		// {"and cannot be used as a key 1/x", "and:value"},
		// {"and cannot be used as a key 2/x", "foo:bar and:value"},
		// {"or cannot be used as a key 1/x", "or:value"},
		// {"or cannot be used as a key 2/x", "coucou:value or:bonmatin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Fatalf(fmt.Sprintf("expected an error but got none | %s", tt.name))
			}
		})
	}
}

func FuzzParser(f *testing.F) {
	f.Fuzz(func(t *testing.T, in string) {
		Parse(in)
	})
}

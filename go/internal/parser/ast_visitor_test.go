package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAST_Parse(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Node
	}{
		{
			name:  "simple field-value pair",
			query: "status:error",
			expected: &TermNode{
				Field:      "status",
				Value:      "error",
				Comparator: Exact,
			},
		},
		{
			name:  "prefix match",
			query: "status:error*",
			expected: &TermNode{
				Field:      "status",
				Value:      "error",
				Comparator: Prefix,
			},
		},
		{
			name:  "suffix match",
			query: "status:*error",
			expected: &TermNode{
				Field:      "status",
				Value:      "error",
				Comparator: Suffix,
			},
		},
		{
			name:  "contains match",
			query: "status:*error*",
			expected: &TermNode{
				Field:      "status",
				Value:      "error",
				Comparator: Contains,
			},
		},
		{
			name:  "AND operation",
			query: "status:error AND level:debug",
			expected: &BooleanNode{
				Left: &TermNode{
					Field:      "status",
					Value:      "error",
					Comparator: Exact,
				},
				Right: &TermNode{
					Field:      "level",
					Value:      "debug",
					Comparator: Exact,
				},
				Kind: And,
			},
		},
		{
			name:  "implicit AND operation",
			query: "status:error level:debug",
			expected: &BooleanNode{
				Left: &TermNode{
					Field:      "status",
					Value:      "error",
					Comparator: Exact,
				},
				Right: &TermNode{
					Field:      "level",
					Value:      "debug",
					Comparator: Exact,
				},
				Kind: And,
			},
		},
		{
			name:  "OR operation",
			query: "status:error OR status:warn",
			expected: &BooleanNode{
				Left: &TermNode{
					Field:      "status",
					Value:      "error",
					Comparator: Exact,
				},
				Right: &TermNode{
					Field:      "status",
					Value:      "warn",
					Comparator: Exact,
				},
				Kind: Or,
			},
		},
		{
			name:  "parenthesized query",
			query: "(status:error OR status:warn) AND level:debug",
			expected: &BooleanNode{
				Left: &BooleanNode{
					Left: &TermNode{
						Field:      "status",
						Value:      "error",
						Comparator: Exact,
					},
					Right: &TermNode{
						Field:      "status",
						Value:      "warn",
						Comparator: Exact,
					},
					Kind: Or,
				},
				Right: &TermNode{
					Field:      "level",
					Value:      "debug",
					Comparator: Exact,
				},
				Kind: And,
			},
		},
		{
			name:  "negation",
			query: "NOT status:error",
			expected: &UnaryBooleanNode{
				Expr: &TermNode{
					Field:      "status",
					Value:      "error",
					Comparator: Exact,
				},
				Kind: Not,
			},
		},
		{
			name:  "AND before OR",
			query: "a:1 AND b:2 OR c:3",
			// (a:1 AND b:2) OR c:3
			expected: &BooleanNode{
				Left: &BooleanNode{
					Left: &TermNode{
						Field:      "a",
						Value:      "1",
						Comparator: Exact,
					},
					Right: &TermNode{
						Field:      "b",
						Value:      "2",
						Comparator: Exact,
					},
					Kind: And,
				},
				Right: &TermNode{
					Field:      "c",
					Value:      "3",
					Comparator: Exact,
				},
				Kind: Or,
			},
		},
		{
			name:  "OR before AND",
			query: "a:1 OR b:2 AND c:3",
			// a:1 OR (b:2 AND c:3)
			expected: &BooleanNode{
				Left: &TermNode{
					Field:      "a",
					Value:      "1",
					Comparator: Exact,
				},
				Right: &BooleanNode{
					Left: &TermNode{
						Field:      "b",
						Value:      "2",
						Comparator: Exact,
					},
					Right: &TermNode{
						Field:      "c",
						Value:      "3",
						Comparator: Exact,
					},
					Kind: And,
				},
				Kind: Or,
			},
		},
		{
			name:  "multiple AND operators",
			query: "a:1 AND b:2 AND c:3",
			// (a:1 AND b:2) AND c:3
			expected: &BooleanNode{
				Left: &BooleanNode{
					Left: &TermNode{
						Field:      "a",
						Value:      "1",
						Comparator: Exact,
					},
					Right: &TermNode{
						Field:      "b",
						Value:      "2",
						Comparator: Exact,
					},
					Kind: And,
				},
				Right: &TermNode{
					Field:      "c",
					Value:      "3",
					Comparator: Exact,
				},
				Kind: And,
			},
		},
		{
			name:  "multiple OR operators",
			query: "a:1 OR b:2 OR c:3",
			// (a:1 OR b:2) OR c:3
			expected: &BooleanNode{
				Left: &BooleanNode{
					Left: &TermNode{
						Field:      "a",
						Value:      "1",
						Comparator: Exact,
					},
					Right: &TermNode{
						Field:      "b",
						Value:      "2",
						Comparator: Exact,
					},
					Kind: Or,
				},
				Right: &TermNode{
					Field:      "c",
					Value:      "3",
					Comparator: Exact,
				},
				Kind: Or,
			},
		},
		{
			name:  "NOT before AND",
			query: "NOT a:1 AND b:2",
			// (NOT a:1) AND b:2
			expected: &BooleanNode{
				Left: &UnaryBooleanNode{
					Expr: &TermNode{
						Field:      "a",
						Value:      "1",
						Comparator: Exact,
					},
					Kind: Not,
				},
				Right: &TermNode{
					Field:      "b",
					Value:      "2",
					Comparator: Exact,
				},
				Kind: And,
			},
		},
		{
			name:  "NOT before OR",
			query: "NOT a:1 OR b:2",
			// (NOT a:1) OR b:2
			expected: &BooleanNode{
				Left: &UnaryBooleanNode{
					Expr: &TermNode{
						Field:      "a",
						Value:      "1",
						Comparator: Exact,
					},
					Kind: Not,
				},
				Right: &TermNode{
					Field:      "b",
					Value:      "2",
					Comparator: Exact,
				},
				Kind: Or,
			},
		},
		{
			name:  "parentheses override precedence",
			query: "a:1 AND (b:2 OR c:3)",
			// a:1 AND (b:2 OR c:3)
			expected: &BooleanNode{
				Left: &TermNode{
					Field:      "a",
					Value:      "1",
					Comparator: Exact,
				},
				Right: &BooleanNode{
					Left: &TermNode{
						Field:      "b",
						Value:      "2",
						Comparator: Exact,
					},
					Right: &TermNode{
						Field:      "c",
						Value:      "3",
						Comparator: Exact,
					},
					Kind: Or,
				},
				Kind: And,
			},
		},
		{
			name:  "complex precedence with NOT",
			query: "a:1 OR NOT b:2 AND c:3",
			// a:1 OR ((NOT b:2) AND c:3)
			expected: &BooleanNode{
				Left: &TermNode{
					Field:      "a",
					Value:      "1",
					Comparator: Exact,
				},
				Right: &BooleanNode{
					Left: &UnaryBooleanNode{
						Expr: &TermNode{
							Field:      "b",
							Value:      "2",
							Comparator: Exact,
						},
						Kind: Not,
					},
					Right: &TermNode{
						Field:      "c",
						Value:      "3",
						Comparator: Exact,
					},
					Kind: And,
				},
				Kind: Or,
			},
		},
		{
			name:  "implicit AND with OR",
			query: "a:1 b:2 OR c:3",
			// (a:1 AND b:2) OR c:3
			expected: &BooleanNode{
				Left: &BooleanNode{
					Left: &TermNode{
						Field:      "a",
						Value:      "1",
						Comparator: Exact,
					},
					Right: &TermNode{
						Field:      "b",
						Value:      "2",
						Comparator: Exact,
					},
					Kind: And,
				},
				Right: &TermNode{
					Field:      "c",
					Value:      "3",
					Comparator: Exact,
				},
				Kind: Or,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, ast.Node)
		})
	}
}

func TestASTBuilder_BuildASTFromQuery_Errors(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "empty query",
			query: "",
		},
		{
			name:  "invalid query",
			query: "status:error AND",
		},
		{
			name:  "missing key",
			query: ":value",
		},
		{
			name:  "missing \":\" separator",
			query: "keywithoutvalue",
		},
		{
			name:  "missing value 1",
			query: "key:",
		},
		{
			name:  "missing value with space",
			query: "key:  ",
		},
		{
			name:  "missing AND second term",
			query: "k:v AND",
		},
		{
			name:  "missing OR second term",
			query: "k:v OR",
		},
		{
			name:  "missing AND first term",
			query: "AND k:v",
		},
		{
			name:  "missing OR first term",
			query: "OR k:v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			assert.Error(t, err)
		})
	}
}

func FuzzParser(f *testing.F) {
	f.Fuzz(func(t *testing.T, in string) {
		Parse(in)
	})
}

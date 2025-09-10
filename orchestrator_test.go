package orchestrator

import (
	"testing"
)

func TestChangesFromUnknownValues(t *testing.T) {
	type Expected = []string

	type Test struct {
		title    string
		values   []any
		expected Expected
	}

	var tests = []Test{
		{
			title: "invalid simple casts",
			values: []any{
				"mockstr",
				"mockstr2",
				int(0),
			},
			expected: nil,
		},
		{
			title: "invalid advanced casts",
			values: []any{
				"mockstr",
				"mockstr2",
				[]any{
					int(0),
					simpleChange{"mockstr3"},
				},
			},
			expected: nil,
		},
		{
			title: "valid simple casts",
			values: []any{
				"mockstr",
				"mockstr2",
				[]string{
					"mockstr3",
					"mockstr4",
				},
			},
			expected: []string{
				"mockstr",
				"mockstr2",
				"mockstr3",
				"mockstr4",
			},
		},
		{
			title: "valid advanced casts",
			values: []any{
				"mockstr",
				"mockstr2",
				[]string{
					"mockstr3",
					"mockstr4",
				},
				simpleChange{"mockstr5"},
				simpleChange{"mockstr6"},
				[]Change{
					simpleChange{"mockstr7"},
					simpleChange{"mockstr8"},
				},
				[]simpleChange{
					// There's a bug here with gofmt, and the only way to fix it is to disable the linting for this specific line
					// nolint:gofmt
					simpleChange{"mockstr9"},
					simpleChange{"mockstr10"},
				},
				[]any{
					simpleChange{"mockstr11"},
				},
			},
			expected: []string{
				"mockstr",
				"mockstr2",
				"mockstr3",
				"mockstr4",
				"mockstr5",
				"mockstr6",
				"mockstr7",
				"mockstr8",
				"mockstr9",
				"mockstr10",
				"mockstr11",
			},
		},
	}

	for _, test := range tests {
		tmp := test
		t.Run(tmp.title, func(t *testing.T) {
			t.Parallel()

			changes, _ := changesFromUnknownValues(tmp.values)
			if len(tmp.expected) != len(changes) {
				t.Fatalf("total number of changes does not match expected value\ngot: %d\nwant: %d", len(changes), len(tmp.expected))
			}

			for i, change := range changes {
				expectedStr := tmp.expected[i]
				changeStr := change.String()
				if expectedStr != changeStr {
					t.Fatalf("change string at index %d does not match expected value\ngot: %s\nwant: %s", i, changeStr, expectedStr)
				}
			}
		})
	}
}

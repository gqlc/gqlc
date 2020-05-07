package cmd

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gqlc/gqlc/gen"
)

func newMockGenerator(t gomock.TestReporter) *gen.MockGenerator {
	return gen.NewMockGenerator(gomock.NewController(t))
}

func TestCli_Run(t *testing.T) {
	c := NewCLI(WithFS(testFs))

	testCases := []struct {
		Name   string
		Args   []string
		expect func(g *gen.MockGenerator)
	}{
		{
			Name: "SingleWoImports",
			Args: []string{"gqlc", "/home/graphql/imports/thr.gql"},
			expect: func(g *gen.MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name: "SingleWImports",
			Args: []string{"gqlc", "-I", "/usr/imports", "-I", "/home/graphql/imports", "five.gql"},
			expect: func(g *gen.MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name: "MultiWoImports",
			Args: []string{"gqlc", "-I", "/home", "-I", "/home/graphql/imports", "thr.gql", "four.gql"},
			expect: func(g *gen.MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name: "MultiWImports",
			Args: []string{"gqlc", "-I", "/usr/imports", "-I", "/home", "-I=/home/graphql", "-I", "/home/graphql/imports", "one.gql", "five.gql"},
			expect: func(g *gen.MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			if err := c.Run(testCase.Args); err != nil {
				subT.Error(err)
				return
			}
		})
	}
}

func compare(t *testing.T, out, ex map[string]interface{}) {
	var match bool
	var missing []string
	for k, outVal := range out {
		exVal, exists := ex[k]
		if !exists {
			missing = append(missing, k)
		}

		switch v := outVal.(type) {
		case int64, float64, string, bool:
			match = v == exVal
		case []int64:
			_, match = exVal.([]int64)
			if !match {
				break
			}

			exSlice := exVal.([]int64)
			match = len(exSlice) == len(v)
			if !match {
				break
			}

			for i := range exSlice {
				if match = v[i] == exSlice[i]; !match {
					break
				}
			}
		case []float64:
			_, match = exVal.([]float64)
			if !match {
				break
			}

			exSlice := exVal.([]float64)
			match = len(exSlice) == len(v)
			if !match {
				break
			}

			for i := range exSlice {
				if match = v[i] == exSlice[i]; !match {
					break
				}
			}
		case []string:
			_, match = exVal.([]string)
			if !match {
				break
			}

			exSlice := exVal.([]string)
			match = len(exSlice) == len(v)
			if !match {
				break
			}

			for i := range exSlice {
				if match = v[i] == exSlice[i]; !match {
					break
				}
			}
		case []bool:
			_, match = exVal.([]bool)
			if !match {
				break
			}

			exSlice := exVal.([]bool)
			match = len(exSlice) == len(v)
			if !match {
				break
			}

			for i := range exSlice {
				if match = v[i] == exSlice[i]; !match {
					break
				}
			}
		default:
			match = false
		}

		if !match {
			t.Fail()
			t.Logf("mismatched values for key, %s: %v:%v", k, outVal, exVal)
		}

		delete(ex, k)
	}

	for _, k := range missing {
		t.Logf("key found in output and not in expected: %s", k)
	}

	for k := range ex {
		t.Logf("expected key: %s", k)
	}
}

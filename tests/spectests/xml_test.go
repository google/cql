// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spectests_test

import (
	"context"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/cql"
	"github.com/google/cql/result"
	"github.com/google/cql/tests/spectests/third_party/cqltests"
	"github.com/google/cql/tests/spectests/exclusions"
	"github.com/google/cql/tests/spectests/models"
	"github.com/google/go-cmp/cmp"
	"slices"
)

// TestCQLXML runs the CQL XML tests by building want and got CQL expression definitions.
// After evaluation, the want and got expression definition result.Values are compared using
// result.Value.Equal. For example:
//
//	<test name="TrueXorTrue">
//		<expression>true xor true</expression>
//		<output>false</output>
//	</test>
//
// Would be translated into:
//
// library CQL_Test
// define "got": true xor true
// define "want": false
//
// After evaluation, the got and want expression definition results are extracted and compared.
func TestCQLXML(t *testing.T) {
	testDir := "."
	testExclusions := exclusions.XMLTestFileExclusionDefinitions()

	files, err := cqltests.XMLTests.ReadDir(testDir)
	if err != nil {
		t.Fatalf("Failed to read cql directory: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("No xml files found in %v", cqltests.XMLTests)
	}

	xmlTestFiles := []string{}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".xml") {
			continue
		}
		xmlTestFiles = append(xmlTestFiles, f.Name())
	}
	if len(xmlTestFiles) == 0 {
		t.Fatal("No xml files found in embedded directory to test")
	}

	for _, fileName := range xmlTestFiles {
		src := filepath.Join(testDir, fileName)
		data, err := cqltests.XMLTests.ReadFile(src)
		if err != nil {
			t.Fatalf("Failed to read XML file: %v", err)
		}

		var cqlTests []cqlTest
		xmlTests, err := parseXML(t, data)
		if err != nil {
			t.Fatalf("failed to parse XML file %s: %v", fileName, err)
		}
		cqlTests = createCQLTests(t, fileName, xmlTests)

		currExclusions, ok := testExclusions[fileName]
		if !ok {
			currExclusions = exclusions.XMLTestFileExclusions{GroupExcludes: []string{}, NamesExcludes: []string{}}
		}

		for _, tc := range cqlTests {
			t.Run(fmt.Sprintf("%s_%s_%s", tc.FileName, tc.Group, tc.Name), func(t *testing.T) {
				shouldSkip := (slices.Contains(currExclusions.GroupExcludes, tc.Group) ||
					slices.Contains(currExclusions.NamesExcludes, tc.Name) ||
					tc.Skip)
				if shouldSkip && strings.Contains(tc.SkipReason, "no output defined") {
					t.Skip("in skipped test groups or names")
				}

				elm, err := cql.Parse(context.Background(), []string{tc.CQL}, cql.ParseConfig{})
				if err != nil {
					if shouldSkip {
						t.Skip("in skipped test groups")
					}
					t.Fatalf("Failed to parse test case: %s\nFilePath=%s\nGroup=%s\nName=%s", err.Error(), src, tc.Group, tc.Name)
				}
				results, err := elm.Eval(context.Background(), nil, cql.EvalConfig{EvaluationTimestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("Fixed", 4*60*60))})
				if err != nil {
					if shouldSkip {
						t.Skip("in skipped test groups")
					}
					t.Fatalf("Failed to execute query(%s): %v\nFilePath=%s\nGroup=%s\nName=%s", tc.CQL, err, src, tc.Group, tc.Name)
				}

				want := getExpDef(t, results, wantExpDef)
				got := getExpDef(t, results, gotExpDef)

				if !cmp.Equal(want, got) {
					if shouldSkip {
						t.Skip("in skipped test groups")
					}
					t.Errorf("evaluating test case %s returned unexpected result. got: %v, want: %v", tc.Name, got, want)
				}
				if shouldSkip {
					t.Errorf("test case %s in group %s was marked to skip but succeeded. You may need to update the test exclusions file", tc.Name, tc.Group)
				}
			})
		}
	}
}

func parseXML(t *testing.T, raw []byte) (models.Tests, error) {
	t.Helper()
	var testCase models.Tests
	if err := xml.Unmarshal(raw, &testCase); err != nil {
		return models.Tests{}, err
	}

	return testCase, nil
}

var (
	gotExpDef  = "got"
	wantExpDef = "want"
)

type cqlTest struct {
	FileName   string
	Group      string
	Name       string
	CQL        string
	Skip       bool
	SkipReason string
}

func createCQLTests(t *testing.T, fileName string, test models.Tests) []cqlTest {
	t.Helper()
	cqlTests := []cqlTest{}
	for _, g := range test.Group {
		for _, tc := range g.Test {
			newTest := cqlTest{FileName: fileName, Group: g.Name, Name: tc.Name}
			if len(tc.Output) == 0 {
				newTest.Skip = true
				newTest.SkipReason = "no output defined for this test case"
			} else {
				cql := fmt.Sprintf("library CQL_Test")
				cql += fmt.Sprintf("\ndefine \"%s\": %s", gotExpDef, tc.Expression.Text)
				cql += fmt.Sprintf("\ndefine \"%s\": %s", wantExpDef, tc.Output[0].Text)
				newTest.CQL = cql
			}
			cqlTests = append(cqlTests, newTest)
		}
	}
	return cqlTests
}

func getExpDef(t *testing.T, results result.Libraries, expDefName string) result.Value {
	libKey := result.LibKey{Name: "CQL_Test"}
	gotExpDefs, ok := results[libKey]
	if !ok {
		t.Fatalf("Failed to find CQL_Test library in CQL output")
	}

	for expDef, gotResult := range gotExpDefs {
		if expDef == expDefName {
			return gotResult
		}
	}
	t.Fatalf("Failed to find expDef %v in CQL output", expDefName)
	return result.Value{}
}

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

// XML Analyzer is a CLI for analyzing engine capabilities vs the external XML tests.
// This is a temporary tool which could eventually be converted to a PresubmitService but for now
// is just a CLI which appends it's findings to the CL argument.
//
package main

import (
	"context"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/cql/tests/spectests/third_party/cqltests"
	"github.com/google/cql/tests/spectests/exclusions"
	"github.com/google/cql/tests/spectests/models"
	"github.com/lithammer/dedent"
)

type cliConfig struct {
}

func (cfg *cliConfig) RegisterFlags(fs *flag.FlagSet) {
}

const usageMessage = "A CLI for analyzing the CQL engine's ability to run the external XML tests."

var errMissingFlag = errors.New("missing required flag")

// The config which is populated by the CLI input flags.
var config cliConfig

func init() {
	config.RegisterFlags(flag.CommandLine)
	defaultUsage := flag.Usage
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageMessage)
		defaultUsage()
	}
}

func main() {
	flag.Parse()

	ctx := context.Background()
	if err := mainWrapper(ctx, config); err != nil {
		log.Fatalf("CQL XML Analyzer failed with an error: %v", err)
	}
}

func mainWrapper(ctx context.Context, cfg cliConfig) error {
	message, err := compileXMLTestStats()
	if err != nil {
		return err
	}
	log.Println(message)

	return nil
}

// compileXMLTestStats compiles the XML test stats from the current workspace.
// Uses the exclusions module to label the tests as excluded or not and then build the stats.
// TODO: b/342065376 - Add a breakdown of the tests that were skipped by file and by groupings.
func compileXMLTestStats() (string, error) {
	testDir := "."
	testExclusions := exclusions.XMLTestFileExclusionDefinitions()

	files, err := cqltests.XMLTests.ReadDir(testDir)
	if err != nil {
		return "", fmt.Errorf("failed to read cql directory: %v", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no xml files found in %s, %v", testDir, cqltests.XMLTests)
	}

	totalMetrics := testMetrics{Name: "Totals"}
	perFileMetrics := map[string]*testMetrics{}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".xml") {
			continue
		}
		src := filepath.Join(testDir, f.Name())
		data, err := cqltests.XMLTests.ReadFile(src)
		if err != nil {
			return "", fmt.Errorf("failed to read XML file: %v", err)
		}

		var cqlTests []cqlTest
		xmlTests, err := parseXML(data)
		if err != nil {
			return "", fmt.Errorf("failed to parse XML file %s: %v", f.Name(), err)
		}
		cqlTests = createCQLTests(xmlTests)

		perFileMetrics[f.Name()] = &testMetrics{Name: f.Name()}
		currExclusions, ok := testExclusions[f.Name()]
		if !ok {
			currExclusions = exclusions.XMLTestFileExclusions{GroupExcludes: []string{}, NamesExcludes: []string{}}
		}

		for _, tc := range cqlTests {
			if slices.Contains(currExclusions.GroupExcludes, tc.Group) {
				totalMetrics.SkippedTests++
				perFileMetrics[f.Name()].SkippedTests++
				continue
			}
			if slices.Contains(currExclusions.NamesExcludes, tc.Name) {
				totalMetrics.SkippedTests++
				perFileMetrics[f.Name()].SkippedTests++
				continue
			}
			if tc.Skip {
				totalMetrics.SkippedTests++
				perFileMetrics[f.Name()].SkippedTests++
				continue
			}
			totalMetrics.ValidTests++
			perFileMetrics[f.Name()].ValidTests++
		}
	}

	r := fmt.Sprintf(`
	CQL XML Text stats:
	%s
	
	====================================

	Per File Stats:`, totalMetrics.toString())
	for _, m := range perFileMetrics {
		r += "\n" + m.toString()
	}
	return dedent.Dedent(r), nil
}

type cqlTest struct {
	Group      string
	Name       string
	Skip       bool
	SkipReason string
}

type testMetrics struct {
	Name         string
	SkippedTests int
	ValidTests   int
}

func (t testMetrics) percentageSkipped() float64 {
	return (float64(t.SkippedTests) / float64(t.totalTests())) * 100.0
}

func (t testMetrics) percentageValid() float64 {
	return (float64(t.ValidTests) / float64(t.totalTests())) * 100.0
}

func (t testMetrics) totalTests() int {
	return t.SkippedTests + t.ValidTests
}

func (t testMetrics) toString() string {
	return fmt.Sprintf(`
	%s:
	%d Skipped (%.2f%%)
	%d Not Skipped (%.2f%%)
	%d Total (%.2f%%)`, t.Name, t.SkippedTests, t.percentageSkipped(), t.ValidTests, t.percentageValid(), t.totalTests(), 100.0)
}

func parseXML(raw []byte) (models.Tests, error) {
	var testCase models.Tests
	if err := xml.Unmarshal(raw, &testCase); err != nil {
		return models.Tests{}, err
	}

	return testCase, nil
}

func createCQLTests(test models.Tests) []cqlTest {
	cqlTests := []cqlTest{}
	for _, g := range test.Group {
		for _, tc := range g.Test {
			newTest := cqlTest{Group: g.Name, Name: tc.Name}
			if len(tc.Output) == 0 {
				newTest.Skip = true
				newTest.SkipReason = "no output defined for this test case"
			}
			cqlTests = append(cqlTests, newTest)
		}
	}
	return cqlTests
}

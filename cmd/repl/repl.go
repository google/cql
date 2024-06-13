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

// The repl is a tool that can be used to interact with the CQL engine iteratively.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"flag"
	"github.com/google/cql"
	"github.com/google/cql/result"
	"github.com/google/cql/retriever/local"
	"github.com/google/cql/terminology"
)

var (
	// In the future we could add an optional alternative flag taking in a directory.
	// This would enable running against multiple resources but may need some work to have
	// a readable output.
	bundleFile = flag.String("bundle_file", "",
		"Path to a single bundle file to seed into the REPL.")
	valuesetsDir = flag.String("valuesets_dir", "",
		"Directory containing JSON versions of FHIR valuesets.")
	cqlFile = flag.String("cql_file", "",
		"Path to a single CQL file to seed into the REPL.")
)

const usageMessage = `This program runs a simple CQL interface for experimenting with the CQL
language in a live environment.

Users can specify an input CQL file, a bundle file resource and a directory
containing FHIR datasets all as inputs to seed data into the REPL ecosystem.

The REPL will then proceed to execute any CQL input text input into the terminal and print
any results. If a line fails to properly parse or execute an error is returned and any inputs are
cleared from the current context.

The '\' character can be used during execution to denote the continuation of an expression
to the next line.

Inputting an empty line is another way of clearing any active context.

Typing 'exit' will exit the program.`

func init() {
	defaultUsage := flag.Usage
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageMessage)
		defaultUsage()
	}
}

// validateFlags checks to ensure that the input flags are valid.
func validateFlags(cqlFile, bundleFile, valuesetsDir string) error {
	// Validate the required bundle file.
	if bundleFile != "" && !strings.HasSuffix(bundleFile, ".json") {
		return fmt.Errorf("--bundle_file when specified, is required to be a valid json file path, check that the input path is valid")
	}

	// Validate the optional CQL input file
	if cqlFile != "" && !strings.HasSuffix(cqlFile, ".cql") {
		return fmt.Errorf("--cql_file flag is required to be a valid .cql file path if provided, check that the input path is valid")
	}

	// Validate the optional valueset directory is valid.
	if valuesetsDir != "" {
		if _, err := os.ReadDir(valuesetsDir); err != nil {
			return err
		}
	}

	return nil
}

// runREPL executes the REPL loop.
// When reading user input `exit` will break from the loop, empty lines
// will clear any current cql context, and lines ending with `\` will be
// treated as continued lines and the current context will be continued
// on through the next line.
func runREPL(seedCQLLibs []string, bundleFileRetriever *local.Retriever, tp terminology.Provider, prevResults result.Libraries) {
	in := bufio.NewReader(os.Stdin)
	var allText, currExpr string

	for {
		if currExpr == "" {
			fmt.Print("> ")
		} else {
			fmt.Print(">> ")
		}

		input, err := in.ReadString('\n')
		if err != nil {
			fmt.Printf("Failed to read input: %v\n", err)
			continue
		}
		input = strings.TrimSpace(input)
		if input == "exit" {
			fmt.Println("Exiting...")
			return
		}
		if input == "list_defs" {
			var defsList []string
			for libKey, libResult := range prevResults {
				for defKey := range libResult {
					defsList = append(defsList, libKey.Name+"."+defKey)
				}
			}
			fmt.Println(strings.Join(defsList, "\n"))
			continue
		}
		if input == "" {
			// If the current expression cache still has data but an empty string
			// was entered, clear the cache and try again.
			if currExpr != "" {
				fmt.Printf("Failed to parse expression text: \n%s\n", currExpr)
				currExpr = ""
			}
			continue
		}
		if strings.HasSuffix(input, `\`) {
			input = strings.TrimSpace(input[:len(input)-1])
			currExpr += input
			continue
		}

		currExpr += input
		currCQL := append(seedCQLLibs, allText+currExpr)
		evalResults, err := runCQLEngine(currCQL, bundleFileRetriever, tp)
		if err != nil {
			fmt.Println(err)
			currExpr = ""
			continue
		}
		resultsText, err := evalResultsDelta(evalResults, prevResults)
		if err != nil {
			fmt.Print(err)
			continue
		}
		// Previous expression evaluated, reset vars and output results.
		prevResults = evalResults
		allText += currExpr + "\n"
		currExpr = ""
		fmt.Println(resultsText)
	}
}

// runCQLEngine runs the CQL engine and returns the results of that execution or error.
func runCQLEngine(cqlLibs []string, bundleFileRetriever *local.Retriever, tp terminology.Provider) (result.Libraries, error) {
	fhirDM, err := cql.FHIRDataModel("4.0.1")
	if err != nil {
		log.Fatal(err)
	}
	elm, err := cql.Parse(context.Background(), cqlLibs, cql.ParseConfig{DataModels: [][]byte{fhirDM}})
	if err != nil {
		// Failed to parse. This case returns error for already defined identifiers
		// and all other parsing related errors.
		return nil, err
	}

	results, err := elm.Eval(context.Background(), bundleFileRetriever, cql.EvalConfig{ReturnPrivateDefs: true, Terminology: tp})
	if err != nil {
		// Failed during execution.
		return nil, err
	}
	return results, nil
}

// evalResultsDelta returns the delta between the previous and current results.
//
// Parse the results and return the json representation of the evaluation results.
// For now naively use the previous results as way to parse out which evaluation results
// do not need to be emitted again.
// In the future we may look into ways to iteratively evaluate cql and only return the
// resulting new results rather than re-evaluating everything.
// Currently there is no way to order the outputs based on the declaration order of input statements.
func evalResultsDelta(evalResults, previousResults result.Libraries) (string, error) {
	var resultDelta []string
	for libKey, libResult := range evalResults {
		for defKey, defObj := range libResult {
			// We previously output the result of this definition.
			if _, prevHasDefKey := previousResults[libKey][defKey]; prevHasDefKey {
				continue
			}

			jr, err := json.Marshal(defObj)
			if err != nil {
				return "", err
			}
			resultDelta = append(resultDelta, string(jr))
		}
	}
	return strings.Join(resultDelta, "\n"), nil
}

func main() {
	flag.Parse()
	if err := validateFlags(*cqlFile, *bundleFile, *valuesetsDir); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Welcome to the CQL REPL! Type exit to leave.")

	retriever := &local.Retriever{}
	if *bundleFile != "" {
		fileData, err := os.ReadFile(*bundleFile)
		if err != nil {
			log.Printf("Error: failed to read input bundle file with error, %v", err)
		}
		retriever, err = local.NewRetrieverFromR4Bundle(fileData)
		if err != nil {
			log.Printf("Error: failed to initialize the bundle retriever with error, %v", err)
		}
	}

	var tp *terminology.LocalFHIRProvider
	if *valuesetsDir != "" {
		var err error
		if tp, err = terminology.NewLocalFHIRProvider(*valuesetsDir); err != nil {
			log.Fatal(err)
		}
	}

	var cqlInput []string
	evalResults := result.Libraries{}
	if *cqlFile != "" {
		bytes, err := os.ReadFile(*cqlFile)
		if err != nil {
			log.Printf("Error: failed to read input CQL file with error, %v", err)
		}
		cqlText := string(bytes)
		if cqlText != "" {
			cqlInput = append(cqlInput, cqlText)
		}
		evalResults, err = runCQLEngine(cqlInput, retriever, tp)
		if err != nil {
			fmt.Println("Ran into error while evaluating input CQL quitting.")
			log.Fatal(err)
		}
		resultsText, err := evalResultsDelta(evalResults, result.Libraries{})
		if err != nil {
			fmt.Print(err)
		}
		fmt.Println(resultsText)
	}

	runREPL(cqlInput, retriever, tp, evalResults)
}

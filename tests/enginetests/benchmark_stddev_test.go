package enginetests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
)

func BenchmarkPopulationStdDev(b *testing.B) {
	// Create a large list of decimals
	count := 10000
	values := make([]string, count)
	for i := 0; i < count; i++ {
		values[i] = fmt.Sprintf("%d.0", i%100)
	}
	listCQL := "{" + strings.Join(values, ", ") + "}"
	cql := "PopulationStdDev(" + listCQL + ")"

	p := newFHIRParser(b)
	parsedLibs, err := p.Libraries(context.Background(), wrapInLib(b, cql), parser.Config{})
	if err != nil {
		b.Fatalf("Parse Libraries returned unexpected error: %v", err)
	}

	config := interpreter.Config{
		DataModels:          p.DataModel(),
		Retriever:           BuildRetriever(b),
		Terminology:         buildTerminologyProvider(b),
		EvaluationTimestamp: defaultEvalTimestamp,
		ReturnPrivateDefs:   true,
	}

	b.ResetTimer()
	b.Run("LargeList", func(b *testing.B) {
		var force result.Libraries
		for n := 0; n < b.N; n++ {
			force, err = interpreter.Eval(context.Background(), parsedLibs, config)
			if err != nil {
				b.Fatalf("Eval returned unexpected error: %v", err)
			}
		}
		forceBenchResult = force
	})
}

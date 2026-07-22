package enginetests

import (
	"context"
	"testing"

	"github.com/google/cql/interpreter"
	"github.com/google/cql/parser"
	"github.com/google/cql/result"
	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCrossLibraryAliasResolution(t *testing.T) {
	tests := []struct {
		name       string
		libraries  []string
		wantResult result.Value
		wantError  string
	}{
		{
			name: "Cross library function calls with alias M - should reproduce the error",
			libraries: []string{
				// Helper library (like CDS_Connect_Commons)
				dedent.Dedent(`
				library TestCommons version '1.0.0'
				using FHIR version '4.0.1'

				define function ActiveMedicationStatement(MedList List<Integer>):
				  MedList M
				    where M > 5

				define function ActiveMedicationRequest(MedList List<Integer>):
				  MedList M
				    where M > 3
				`),
				// Main library (like Statin Therapy)
				dedent.Dedent(`
				library MainLib version '1.0.0'
				using FHIR version '4.0.1'
				include TestCommons version '1.0.0' called TC

				define TESTRESULT: 
				  exists(TC.ActiveMedicationStatement({1, 6, 3, 8}))
				  or exists(TC.ActiveMedicationRequest({1, 2, 4, 7}))
				`),
			},
			wantResult: newOrFatal(t, true),
		},
		{
			name: "Cross library with FHIR retrievals - closer to real scenario",
			libraries: []string{
				// Helper library with FHIR functions
				dedent.Dedent(`
				library FHIRCommons version '1.0.0'
				using FHIR version '4.0.1'

				define function ActiveMedicationStatement(MedList List<MedicationStatement>):
				  MedList M
				    where M.status.value = 'active'

				define function ActiveMedicationRequest(MedList List<MedicationRequest>):
				  MedList M
				    where M.status.value = 'active'
				`),
				// Main library calling FHIR functions
				dedent.Dedent(`
				library StatinLib version '1.0.0'
				using FHIR version '4.0.1'
				include FHIRCommons version '1.0.0' called FC

				define TESTRESULT: 
				  exists(FC.ActiveMedicationStatement([MedicationStatement]))
				  or exists(FC.ActiveMedicationRequest([MedicationRequest]))
				`),
			},
			wantResult: newOrFatal(t, false), // Empty retrievals
		},
		{
			name: "Exact CDS_Connect_Commons pattern with let clause - works with our fix",
			libraries: []string{
				// Exact pattern from CDS_Connect_Commons
				dedent.Dedent(`
				library CDSCommons version '1.0.0'
				using FHIR version '4.0.1'

				define function PeriodToInterval(period FHIR.Period):
				  if period is null then
				    null
				  else
				    Interval[period."start".value, period."end".value]

				define function ActiveMedicationStatement(MedList List<MedicationStatement>):
				  MedList M
				    let EffectivePeriod: PeriodToInterval(M.effective as FHIR.Period)
				    where M.status.value = 'active'
				      and (end of EffectivePeriod is null or end of EffectivePeriod after Now())
				`),
				// Main library that calls the function
				dedent.Dedent(`
				library MainLib version '1.0.0'
				using FHIR version '4.0.1'
				include CDSCommons version '1.0.0' called C3F

				define TESTRESULT: 
				  exists(C3F.ActiveMedicationStatement([MedicationStatement]))
				`),
			},
			wantResult: newOrFatal(t, false), // Empty retrievals
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newFHIRParser(t)
			
			// Add FHIRHelpers to all libraries
			libsWithHelpers := make([]string, len(tc.libraries))
			for i, lib := range tc.libraries {
				libsWithHelpers[i] = addFHIRHelpersLib(t, lib)[0]
			}
			
			parsedLibs, err := p.Libraries(context.Background(), libsWithHelpers, parser.Config{})
			if tc.wantError != "" {
				if err == nil {
					t.Fatalf("Expected parse error containing %q, but got no error", tc.wantError)
				}
				if !contains(err.Error(), tc.wantError) {
					t.Fatalf("Expected parse error containing %q, but got: %v", tc.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			results, err := interpreter.Eval(context.Background(), parsedLibs, defaultInterpreterConfig(t, p))
			if tc.wantError != "" {
				if err == nil {
					t.Fatalf("Expected eval error containing %q, but got no error", tc.wantError)
				}
				if !contains(err.Error(), tc.wantError) {
					t.Fatalf("Expected eval error containing %q, but got: %v", tc.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Eval returned unexpected error: %v", err)
			}
			
			gotResult := getTESTRESULTWithSources(t, results)
			if diff := cmp.Diff(tc.wantResult, gotResult, protocmp.Transform()); diff != "" {
				t.Errorf("Eval diff (-want +got)\n%v", diff)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}

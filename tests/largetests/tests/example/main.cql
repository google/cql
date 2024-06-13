library main version '0.0.1'
using FHIR version '4.0.1'

valueset "Systolic blood pressure": 'valueset-systolic-blood-pressure' version '1.0.0'

define "Valid systolic blood pressures sorted by effective":
 [Observation: "Systolic blood pressure"] bp
      where bp.status.value in {'final', 'amended', 'corrected'}
      sort by effective desc

define "Most recent systolic blood pressure":
  Last("Valid systolic blood pressures sorted by effective")
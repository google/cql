library RadiologyAnomalyDetection version '1.0.0'
using FHIR version '4.0.1'
include FHIRHelpers version '4.0.1'

context Patient

define "RelevantRadiologyReports":
  [DiagnosticReport] DR
    where DR.status.value in {'final', 'amended', 'corrected'}
      and exists (DR.category C where exists (C.coding Coding where Coding.code.value = 'RAD'))

define "FlaggedAnomalies":
  "RelevantRadiologyReports" DR
    where exists (DR.conclusionCode.coding C
      where C.system.value = 'http://snomed.info/sct' and C.code.value = '386661006')

define "FinalDecision":
  if exists("FlaggedAnomalies") then 'Prioritize for Urgent AI-Assisted Review'
  else 'Standard Workflow'

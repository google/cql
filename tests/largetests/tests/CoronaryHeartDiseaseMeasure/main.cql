library M2 version '0.0.1'
using FHIR version '4.0.1'

include FHIRHelpers version '4.0.1' called FHIRHelpers

valueset "Coronary arteriosclerosis": 'url-valueset-coronary-arteriosclerosis' version '1.0.0'
valueset "Systolic blood pressure": 'valueset-systolic-blood-pressure' version '1.0.0'

parameter "Measurement Period" Interval<DateTime>
  default Interval[@2010-04-01T00:00:00.000Z, @2024-03-31T00:00:00.000Z)

context Patient

define "Has coronary heart disease":
  exists (
    [Condition: "Coronary arteriosclerosis"] chd
      where chd.onset as FHIR.dateTime before day of start of "Measurement Period"
  )

define "Most recent blood pressure reading":
  Last(
    [Observation: "Systolic blood pressure"] bp
      where bp.status in {'final', 'amended', 'corrected'}
      and bp.effective in day of "Measurement Period"
      sort by effective desc
  )

define "Most recent blood pressure reading below 150":
   "Most recent blood pressure reading".value < 150

define "Initial Population":
  AgeInYearsAt(start of "Measurement Period") > 80

define "Denominator":
  "Initial Population"
  and "Has coronary heart disease"

define "Numerator":
  "Initial Population"
  and "Denominator"
  and "Most recent blood pressure reading below 150"
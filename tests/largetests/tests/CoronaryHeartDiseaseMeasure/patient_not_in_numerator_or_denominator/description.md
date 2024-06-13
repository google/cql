This test case is a patient with the following attributes:

* Above age 80 by start of measurement period
* Has coronary heart disease but onset _after_ measurement period
* Most recent systolic blood pressure reading is under 150

We expect this patient to be in the initial population, but not the
numerator or denominator.

TODO(b/323418402): allow comments in JSON to include descriptions in the data
file directly.

TODO(b/319503337): support config based data generation options.
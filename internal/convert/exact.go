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

package convert

import (
	"fmt"
	"math"

	"github.com/google/cql/internal/modelinfo"
	"github.com/google/cql/types"
)

// ExactOverloadMatch returns F on a match, and an error if there is no match or if the match is
// ambiguous. Matches must be exact meaning the invoked operands are equal or a subtype of the
// matched overload. If there are two exact matches an ambiguous error is returned.
func ExactOverloadMatch[F any](invoked []types.IType, overloads []Overload[F], modelinfo *modelinfo.ModelInfos, name string) (F, error) {
	if len(overloads) == 0 {
		return zero[F](), fmt.Errorf("could not resolve %v(%v): %w", name, types.ToStrings(invoked), ErrNoMatch)
	}

	foundMatch := false
	minScore := math.MaxInt
	ambiguous := false
	var matched F
	for _, overload := range overloads {
		match, score, err := operandsExactOrSubtypeMatch(invoked, overload.Operands, modelinfo)
		if err != nil {
			return zero[F](), fmt.Errorf("%v(%v): %w", name, types.ToStrings(invoked), err)
		}

		if match && score == minScore {
			// Least converting match is now ambiguous
			ambiguous = true
			continue
		}

		if match && score < minScore {
			// New least converting match
			foundMatch = true
			ambiguous = false
			minScore = score
			matched = overload.Result
		}
	}

	if foundMatch && ambiguous {
		return zero[F](), fmt.Errorf("%v(%v) ambiguous match", name, types.ToStrings(invoked))
	}
	if foundMatch {
		return matched, nil
	}

	return zero[F](), fmt.Errorf("could not resolve %v(%v): %w", name, types.ToStrings(invoked), ErrNoMatch)
}

func operandsExactOrSubtypeMatch(invoked []types.IType, declared []types.IType, modelinfo *modelinfo.ModelInfos) (bool, int, error) {
	if len(invoked) != len(declared) {
		return false, 0, nil
	}

	score := 0
	for i := range invoked {
		if invoked[i] == types.Unset {
			return false, score, fmt.Errorf("internal error - invokedType is types.Unsupported")
		}

		if declared[i] == types.Unset {
			return false, score, fmt.Errorf("internal error - declaredType is types.Unsupported")
		}
		// EXACT MATCH
		if invoked[i].Equal(declared[i]) {
			continue
		}

		// SUBTYPE
		isSub, err := modelinfo.IsSubType(invoked[i], declared[i])
		if err != nil {
			return false, score, err
		}
		if !isSub {
			return false, score, nil
		}
		score++ // is a subtype match, so we increment
	}

	return true, score, nil
}

// zero is a helper function to return the Zero value of a generic type T.
func zero[T any]() T {
	var zero T
	return zero
}

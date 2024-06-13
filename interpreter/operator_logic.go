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

package interpreter

import (
	"fmt"

	"github.com/google/cql/model"
	"github.com/google/cql/result"
)

// LOGICAL OPERATORS - https://cql.hl7.org/09-b-cqlreference.html#logical-operators-3

// and (left Boolean, right Boolean) Boolean
// or (left Boolean, right Boolean) Boolean
// xor (left Boolean, right Boolean) Boolean
// implies (left Boolean, right Boolean) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#and
// https://cql.hl7.org/09-b-cqlreference.html#or
// https://cql.hl7.org/09-b-cqlreference.html#xor
// https://cql.hl7.org/09-b-cqlreference.html#implies
func evalLogic(m model.IBinaryExpression, lObj, rObj result.Value) (result.Value, error) {
	l := new(bool)
	r := new(bool)
	var err error
	if result.IsNull(lObj) {
		l = nil
	} else {
		*l, err = result.ToBool(lObj)
		if err != nil {
			return result.Value{}, err
		}
	}

	if result.IsNull(rObj) {
		r = nil
	} else {
		*r, err = result.ToBool(rObj)
		if err != nil {
			return result.Value{}, err
		}
	}

	switch m.(type) {
	case *model.And:
		return and(l, r)
	case *model.Or:
		return or(l, r)
	case *model.XOr:
		return xor(l, r)
	case *model.Implies:
		return implies(l, r)
	}
	return result.Value{}, fmt.Errorf("Unsupported BinaryLogicExpression %v", m)
}

/*
	         right
	+-------+------+-------+------+
	|       | TRUE | FALSE | NULL |

l +-------+------+-------+------+
e | TRUE  | TRUE | TRUE  | TRUE |
f +-------+------+-------+------+
t | FALSE | TRUE | FALSE | NULL |

	+-------+------+-------+------+
	| NULL  | TRUE | NULL  | NULL |
	+-------+------+-------+------+
*/
func or(l, r *bool) (result.Value, error) {
	if l != nil && *l {
		return result.New(true)
	}
	if r != nil && *r {
		return result.New(true)
	}
	if l == nil || r == nil {
		return result.New(nil)
	}
	return result.New(false)
}

/*
	         right
	+-------+-------+-------+-------+
	|       | TRUE  | FALSE | NULL  |

l +-------+-------+-------+-------+
e | TRUE  | TRUE  | FALSE | NULL  |
f +-------+-------+-------+-------+
t | FALSE | FALSE | FALSE | FALSE |

	+-------+-------+-------+-------+
	| NULL  | NULL  | FALSE | NULL  |
	+-------+-------+-------+-------+
*/
func and(l, r *bool) (result.Value, error) {
	if l != nil && r != nil {
		return result.New(*l && *r)
	}
	if l != nil && !*l {
		return result.New(false)
	}
	if r != nil && !*r {
		return result.New(false)
	}
	return result.New(nil)
}

/*
	         right
	+-------+-------+-------+------+
	|       | TRUE  | FALSE | NULL |

l +-------+-------+-------+------+
e | TRUE  | FALSE | TRUE  | NULL |
f +-------+-------+-------+------+
t | FALSE | TRUE  | FALSE | NULL |

	+-------+-------+-------+------+
	| NULL  | NULL  | NULL  | NULL |
	+-------+-------+-------+------+
*/
func xor(l, r *bool) (result.Value, error) {
	if l == nil || r == nil {
		return result.New(nil)
	}
	return result.New(*l != *r)
}

/*
	         right
	+-------+------+-------+-------+
	|       | TRUE | FALSE | NULL  |

l +-------+------+-------+-------+
e | TRUE  | TRUE | FALSE | NULL  |
f +-------+------+-------+-------+
t | FALSE | TRUE | TRUE  | TRUE  |

	+-------+------+-------+-------+
	| NULL  | TRUE | NULL  | NULL  |
	+-------+------+-------+-------+
*/
func implies(l, r *bool) (result.Value, error) {
	if l == nil {
		if r != nil && *r {
			return result.New(true)
		}
		return result.New(nil)
	}

	if r == nil {
		if !*l {
			return result.New(true)
		}
		return result.New(nil)
	}

	if *r {
		return result.New(true)
	}
	if !*l {
		return result.New(true)
	}

	return result.New(*r)
}

// not (argument Boolean) Boolean
// https://cql.hl7.org/09-b-cqlreference.html#not
func evalNot(m model.IUnaryExpression, obj result.Value) (result.Value, error) {
	if result.IsNull(obj) {
		return result.New(nil)
	}

	boolObj, err := result.ToBool(obj)
	if err != nil {
		return result.Value{}, err
	}
	return result.New(!boolObj)
}

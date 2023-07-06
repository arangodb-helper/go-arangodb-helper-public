//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package errors

import (
	"fmt"

	errs "github.com/pkg/errors"
)

var (
	Cause        = errs.Cause
	New          = errs.New
	WithMessage  = errs.WithMessage
	WithMessagef = errs.WithMessagef
)

// CauseWithNil returns Cause of an error.
// If error returned by Cause is same (no Causer interface implemented), function will return nil instead
func CauseWithNil(err error) error {
	if nerr := Cause(err); err == nil {
		return nil
	} else if nerr == err {
		// Cause returns same error if error object does not implement Causer interface
		// To prevent infinite loops in usage CauseWithNil will return nil in this case
		return nil
	} else {
		return nerr
	}
}

func Newf(format string, args ...interface{}) error {
	return New(fmt.Sprintf(format, args...))
}

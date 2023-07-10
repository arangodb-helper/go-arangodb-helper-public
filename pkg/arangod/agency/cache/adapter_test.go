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

package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb-helper/go-helper/pkg/arangod/tests"
	"github.com/arangodb/go-driver"
)

func TestNewDriverV1DiscoveryAdapter(t *testing.T) {
	s := tests.NewServer(t)

	fn := func(ctx context.Context) (driver.Connection, error) {
		return s.NewConnection(), nil
	}

	adapter := NewDriverV1DiscoveryAdapter(fn)
	c, err := adapter.Discover(context.Background())
	require.NoError(t, err)
	require.NotNil(t, c)
}

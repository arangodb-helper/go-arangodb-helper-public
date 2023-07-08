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
	"bytes"
	"context"
	"io"

	"github.com/arangodb-helper/go-helper/pkg/arangod/conn"
	"github.com/arangodb-helper/go-helper/pkg/errors"
	driver "github.com/arangodb/go-driver"
)

type DriverV1ClientGetter func(ctx context.Context) (driver.Connection, error)

type DriverV1DiscoveryAdapter struct {
	getter DriverV1ClientGetter
}

var _ LeaderDiscovery = &DriverV1DiscoveryAdapter{}

// NewDriverV1DiscoveryAdapter accepts a function which returns driver.Connection and
// returns a compatible LeaderDiscovery implementation
func NewDriverV1DiscoveryAdapter(fn DriverV1ClientGetter) LeaderDiscovery {
	return &DriverV1DiscoveryAdapter{
		getter: fn,
	}
}

func (da *DriverV1DiscoveryAdapter) Discover(ctx context.Context) (conn.Connection, error) {
	connection, err := da.getter(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get client")
	}

	return &driverV1AgencyAdapter{
		c: connection,
	}, nil
}

type driverV1AgencyAdapter struct {
	c driver.Connection
}

func (a *driverV1AgencyAdapter) Execute(ctx context.Context, method string, endpoint string, body io.Reader) (io.ReadCloser, int, error) {
	req, err := a.c.NewRequest(method, endpoint)
	if err != nil {
		return nil, 0, errors.WithMessage(err, "NewRequest failed)")
	}
	requestBody, err := io.ReadAll(body) // driver v1 connection does not support io.Reader as input
	if err != nil {
		return nil, 0, errors.WithMessage(err, "ReadAll failed for request body")
	}
	req, err = req.SetBody(requestBody)
	if err != nil {
		return nil, 0, errors.WithMessage(err, "SetBody failed")
	}

	var rawResponse []byte
	resp, err := a.c.Do(driver.WithRawResponse(ctx, &rawResponse), req)
	if err != nil {
		return nil, 0, errors.WithMessage(err, "Do failed")
	}
	return io.NopCloser(bytes.NewBuffer(rawResponse)), resp.StatusCode(), nil
}

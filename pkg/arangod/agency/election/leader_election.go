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

package election

import (
	"context"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
)

const (
	keyData = "data"
	keyTTL  = "ttl"
)

func NewLeaderElectionCell[T comparable](key []string, ttl time.Duration) *LeaderElectionCell[T] {
	return &LeaderElectionCell[T]{
		lastTTL: 0,
		leading: false,
		key:     key,
		ttl:     ttl,
	}
}

type LeaderElectionCell[T comparable] struct {
	lastTTL int64
	leading bool
	key     []string
	ttl     time.Duration
}

type leaderStruct[T comparable] struct {
	Data T     `json:"data,omitempty"` // keyData
	TTL  int64 `json:"ttl,omitempty"`  // keyTTL
}

// GetLeaderCondition creates a condition which is resolved to true only if current value in agency equals provided dataValue
func (l *LeaderElectionCell[T]) GetLeaderCondition(dataValue T) agency.ConditionsMap {
	return agency.ConditionsMap{
		createFullKey(append(l.key, keyData)): agency.NewConditionIfEqual(dataValue),
	}
}

func (l *LeaderElectionCell[T]) tryBecomeLeader(ctx context.Context, cli agency.Agency, value T, assumeEmpty bool) error {
	trx := agency.NewTransaction("", agency.TransactionOptions{})

	newTTL := time.Now().Add(l.ttl).Unix()
	trx.AddKey(agency.NewKeySet(l.key, leaderStruct[T]{Data: value, TTL: newTTL}, 0))
	if assumeEmpty {
		trx.AddCondition(l.key, agency.NewConditionOldEmpty(true))
	} else {
		key := append(l.key, keyTTL)
		trx.AddCondition(key, agency.NewConditionIfEqual(l.lastTTL))
	}

	if err := cli.WriteTransaction(ctx, trx); err == nil {
		l.lastTTL = newTTL
		l.leading = true
	} else {
		return err
	}

	return nil
}

func (l *LeaderElectionCell[T]) Read(ctx context.Context, cli agency.Agency) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var result leaderStruct[T]
	if err := cli.ReadKey(ctx, l.key, &result); err != nil {
		var def T
		if agency.IsKeyNotFound(err) {
			return def, nil
		}
		return def, err
	}
	return result.Data, nil
}

// Update checks the current leader cell and if no leader is present
// it tries to put itself in there. Will return the value currently present,
// whether we are leader and a duration after which Updated should be called again.
func (l *LeaderElectionCell[T]) Update(ctx context.Context, cli agency.Agency, value T) (T, bool, time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		assumeEmpty := false
		var result leaderStruct[T]
		if err := cli.ReadKey(ctx, l.key, &result); err != nil {
			if agency.IsKeyNotFound(err) {
				assumeEmpty = true
				goto tryLeaderElection
			}
			assumeEmpty = false
		}

		{
			now := time.Now()
			if result.TTL < now.Unix() {
				l.lastTTL = result.TTL
				l.leading = false
				goto tryLeaderElection
			}

			if result.TTL == l.lastTTL && l.leading {
				// try to update the ttl
				goto tryLeaderElection
			} else {
				// some new leader has been established
				l.lastTTL = result.TTL
				l.leading = false
				return result.Data, false, time.Unix(l.lastTTL, 0).Sub(now), nil
			}
		}

	tryLeaderElection:
		if err := l.tryBecomeLeader(ctx, cli, value, assumeEmpty); err == nil {
			return value, true, l.ttl / 2, nil
		} else if !driver.IsPreconditionFailed(err) {
			var def T
			return def, false, 0, err
		}
	}
}

// Resign tries to resign leadership. If error is returned, caller should retry
func (l *LeaderElectionCell[T]) Resign(ctx context.Context, cli agency.Agency) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// delete the key with precondition that ttl is as expected
	if !l.leading {
		return nil
	}
	l.leading = false
	trx := agency.NewTransaction("", agency.TransactionOptions{})
	key := append(l.key, keyTTL)
	trx.AddCondition(key, agency.NewConditionIfEqual(l.lastTTL))
	trx.AddKey(agency.NewKeyDelete(l.key))
	err := cli.WriteTransaction(ctx, trx)
	if err != nil && driver.IsPreconditionFailed(err) {
		//  we're no longer the leader
		return nil
	}
	return err
}

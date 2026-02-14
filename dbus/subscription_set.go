// Copyright 2015 CoreOS, Inc.
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

package dbus

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// SubscriptionSet returns a subscription set which is like conn.Subscribe but
// can filter to only return events for a set of units.
type SubscriptionSet struct {
	*set
	conn *Conn
}

func (s *SubscriptionSet) filter(unit string) bool {
	return !s.Contains(unit)
}

// SubscribeContext starts listening for dbus events for all of the units in the set.
// Returns channels identical to conn.SubscribeUnits.
func (s *SubscriptionSet) SubscribeContext(ctx context.Context) (<-chan map[string]*UnitStatus, <-chan error) {
	return s.conn.SubscribeUnitsCustomContext(ctx, time.Second, 0,
		mismatchUnitStatus,
		func(unit string) bool { return s.filter(unit) },
	)
}

// Deprecated: use SubscribeContext instead.
func (s *SubscriptionSet) Subscribe() (<-chan map[string]*UnitStatus, <-chan error) {
	return s.SubscribeContext(context.Background())
}

// NewSubscriptionSet returns a new subscription set.
func (c *Conn) NewSubscriptionSet() *SubscriptionSet {
	return &SubscriptionSet{newSet(), c}
}

// SetPropertiesSubscriber works the same as [Conn.SetPropertiesSubscriber] but will send updates
// only for the units that are part of [SubscriptionSet].
func (s *SubscriptionSet) SetPropertiesSubscriber(ctx context.Context, propertiesChangedCh chan<- *PropertiesUpdate, errorCh chan<- error) {
	for unit := range s.data {
		s.conn.addMatchUnitPropertiesChanged(unit)
	}

	s.conn.propertiesSubscriber.Lock()
	defer s.conn.propertiesSubscriber.Unlock()
	s.conn.propertiesSubscriber.subscriptionSets = append(s.conn.propertiesSubscriber.subscriptionSets, struct {
		set      *SubscriptionSet
		updateCh chan<- *PropertiesUpdate
		errCh    chan<- error
	}{set: s, updateCh: propertiesChangedCh, errCh: errorCh})

	go func() {
		<-ctx.Done()

		s.conn.propertiesSubscriber.Lock()
		defer s.conn.propertiesSubscriber.Unlock()
		for idx, setWithChannel := range s.conn.propertiesSubscriber.subscriptionSets {
			if s == setWithChannel.set {
				s.conn.propertiesSubscriber.subscriptionSets = append(s.conn.propertiesSubscriber.subscriptionSets[:idx], s.conn.propertiesSubscriber.subscriptionSets[idx+1:]...)
				close(setWithChannel.updateCh)
				close(setWithChannel.errCh)

				for unit := range s.data {
					if !s.conn.unitReferenced(unit) {
						s.conn.removeMatchUnitPropertiesChanged(unit)
					}
				}
				break
			}
		}
	}()
}

func (s *SubscriptionSet) Add(value string) {
	s.set.Add(value)
	s.conn.addMatchUnitPropertiesChanged(value)
}

func (s *SubscriptionSet) Remove(value string) {
	s.set.Remove(value)

	s.conn.propertiesSubscriber.Lock()
	defer s.conn.propertiesSubscriber.Unlock()
	if !s.conn.unitReferenced(value) {
		s.conn.removeMatchUnitPropertiesChanged(value)
	}
}

func (c *Conn) addMatchUnitPropertiesChanged(unit string) {
	c.sigconn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',path='/org/freedesktop/systemd1/unit/%s'", escapeUnitNameForDBus(unit)))
}

func (c *Conn) removeMatchUnitPropertiesChanged(unit string) {
	c.sigconn.BusObject().Call("org.freedesktop.DBus.RemoveMatch", 0,
		fmt.Sprintf("type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',path='/org/freedesktop/systemd1/unit/%s'", escapeUnitNameForDBus(unit)))
}

func (c *Conn) unitReferenced(unit string) bool {
	for _, s := range c.propertiesSubscriber.subscriptionSets {
		_, ok := s.set.data[unit]
		if ok {
			return true
		}
	}

	if c.propertiesSubscriber.unitSubscriptions == nil {
		return false
	}

	_, ok := c.propertiesSubscriber.unitSubscriptions[unit]
	if ok {
		return true
	}
	return false
}

func escapeUnitNameForDBus(unit string) string {
	var escaped strings.Builder
	for _, r := range unit {
		if (unicode.IsLetter(r) || unicode.IsDigit(r)) && r < 128 {
			escaped.WriteRune(r)
		} else {
			escaped.WriteString(fmt.Sprintf("_%02x", r))
		}
	}
	return escaped.String()
}

// mismatchUnitStatus returns true if the provided UnitStatus objects
// are not equivalent. false is returned if the objects are equivalent.
// Only the Name, Description and state-related fields are used in
// the comparison.
func mismatchUnitStatus(u1, u2 *UnitStatus) bool {
	return u1.Name != u2.Name ||
		u1.Description != u2.Description ||
		u1.LoadState != u2.LoadState ||
		u1.ActiveState != u2.ActiveState ||
		u1.SubState != u2.SubState
}

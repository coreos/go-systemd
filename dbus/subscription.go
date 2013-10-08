package dbus

import (
	"time"
)

// Returns two unbuffered channels which will receive all changed units every
// @interval@ seconds.  Deleted units are sent as nil.
func (c *Conn) SubscribeUnits(interval time.Duration) (<-chan map[string]*UnitStatus, <-chan error) {
	return c.SubscribeUnitsCustom(interval, 0, func(u1, u2 *UnitStatus) bool { return *u1 != *u2 })
}

// SubscribeUnitsCustom is like SubscribeUnits but lets you specify the buffer
// size of the channels and the comparison function for detecting changes.
func (c *Conn) SubscribeUnitsCustom(interval time.Duration, buffer int, isChanged func(*UnitStatus, *UnitStatus) bool) (<-chan map[string]*UnitStatus, <-chan error) {
	old := make(map[string]*UnitStatus)
	statusChan := make(chan map[string]*UnitStatus, buffer)
	errChan := make(chan error, buffer)

	go func() {
		for {
			timerChan := time.After(interval)

			units, err := c.ListUnits()
			if err == nil {
				cur := make(map[string]*UnitStatus)
				for i := range units {
					cur[units[i].Name] = &units[i]
				}

				// add all new or changed units
				changed := make(map[string]*UnitStatus)
				for n, u := range cur {
					if oldU, ok := old[n]; !ok || isChanged(oldU, u) {
						changed[n] = u
					}
					delete(old, n)
				}

				// add all deleted units
				for oldN := range old {
					changed[oldN] = nil
				}

				old = cur

				statusChan <- changed
			} else {
				errChan <- err
			}

			<-timerChan
		}
	}()

	return statusChan, errChan
}

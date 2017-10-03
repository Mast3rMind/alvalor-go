// Copyright (c) 2017 The Alvalor Authors
//
// This file is part of Alvalor.
//
// Alvalor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Alvalor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package book

import (
	"errors"
	"math"
	"math/rand"
	"sort"
	"sync"
)

// score returns the score used for sorting entries by priority. The score of entries that have
// never failed is always one. For entries that failed, the score is the ratio between successes
// and failures.
func (e Entry) score() float64 {
	failure := e.Int("failure")
	if failure == 0 {
		return 1
	}
	success := e.Int("success")
	score := float64(success) / float64(failure)
	return math.Min(score/100, 1)
}

// enumeration of different errors that we can return from address book functions.
var (
	errAddrInvalid = errors.New("invalid address")
	errAddrUnknown = errors.New("unknown address")
	errBookEmpty   = errors.New("book empty")
)

// SimpleBook represents an address book that uses naive algorithms to manage peer addresses. It has
// an explicit blacklist, a registry of peers and defines a sample size of how many addresses to
// return on its random sample. It uses a mutex to be thread-safe.
type SimpleBook struct {
	mutex       sync.Mutex
	blacklisted map[string]struct{}
	entries     map[string]*Entry
}

// NewSimpleBook creates a new default initialized instance of a simple address book.
func NewSimpleBook() *SimpleBook {
	return &SimpleBook{
		blacklisted: make(map[string]struct{}),
		entries:     make(map[string]*Entry),
	}
}

// IsActive represents filter to select active/inactive entries in Sample method
func IsActive(active bool) func(e *Entry) bool {
	return func(e *Entry) bool {
		return e.Bool("active") == active
	}
}

// Any reperesents filter to select any entries in Sample method
func Any() func(e *Entry) bool {
	return func(e *Entry) bool {
		return true
	}
}

// ByPrioritySort represents an order by priority which is calculated based on score. It can be used to sort entries in Sample method
func ByPrioritySort() func([]*Entry) []*Entry {
	return func(entries []*Entry) []*Entry {
		sort.Sort(byPriority(entries))
		return entries
	}
}

// RandomSort represents a random order. It can be used to sort entries in Sample method
func RandomSort() func([]*Entry) []*Entry {
	return func(entries []*Entry) []*Entry {
		for i := 0; i < len(entries); i++ {
			j := rand.Intn(i + 1)
			entries[i], entries[j] = entries[j], entries[i]
		}
		return entries
	}
}

// Action executes a given action on the address.
func (s *SimpleBook) Action(address string, action Action) {
	switch action {
	case ActionAdd:
		s.add(address)
	case ActionWhitelist:
		s.whitelist(address)
	case ActionBlacklist:
		s.blacklist(address)
	default:
	}
}

// Whitelist will remove an address from the blacklist and set it's score to one by resetting
// failures and setting success count to one.
func (s *SimpleBook) whitelist(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.blacklisted, address)
	e, ok := s.entries[address]
	if !ok {
		return
	}
	e.Set("failure", 0)
	e.Set("sucess", 1)
}

// Blacklist will delete an entry from the registry and put it in the blacklist so further adding
// is no longer possible.
func (s *SimpleBook) blacklist(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.entries, address)
	s.blacklisted[address] = struct{}{}
}

// Add will add an address to the list of available peer addresses, unless it is blacklisted.
func (s *SimpleBook) add(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.blacklisted[address]
	if ok {
		return
	}
	s.entries[address] = &Entry{Address: address, Properties: make(map[string]interface{})}
}

// Event notifies the address book of an event of the given type.
func (s *SimpleBook) Event(address string, event Event) {
	switch event {
	case EventConnected:
		s.connected(address)
	case EventDisconnected:
		s.disconnected(address)
	case EventDropped:
		s.dropped(address)
	case EventFailed:
		s.failed(address)
	default:
	}
}

// Connected should be called by consumers whenever a successful connection to the peer with the
// given address was established. It is used to keep track of the active status and to increase the
// success count of the peer.
func (s *SimpleBook) connected(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[address]
	if !ok {
		return
	}
	e.Set("active", true)
	e.Set("success", e.Int("success")+1)
}

// Disconnected should be called by consumers whenever a peer was disconnected in a clean way. It is
// used to keep track of the active status.
func (s *SimpleBook) disconnected(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[address]
	if !ok {
		return
	}
	e.Set("active", false)
}

// Dropped should be called by consumers whenever a we decided to drop a peer due to errors in the
// connection. It is used to keep track of the failure & active status of a peer.
func (s *SimpleBook) dropped(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[address]
	if !ok {
		return
	}
	e.Set("active", false)
	e.Set("failure", e.Int("failure")+1)
}

// Failed should be called whenever connection to a peer failed. It is used to keep track of the
// failure & active status of a peer.
func (s *SimpleBook) failed(addr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[addr]
	if !ok {
		return
	}
	e.Set("active", false)
	e.Set("failure", e.Int("failure")+1)
}

// Sample will return entries limited by count, filtered by specified filter function and sorted by specified sort function
func (s *SimpleBook) Sample(filter func(*Entry) bool, sort func([]*Entry) []*Entry, limit uint) ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	entries := s.slice(filter)

	if len(entries) == 0 {
		return nil, errBookEmpty
	}

	entries = sort(entries)

	if len(entries) > int(limit) {
		entries = entries[:limit]
	}

	addresses := make([]string, 0, limit)
	for _, e := range entries {
		addresses = append(addresses, e.Address)
	}
	return addresses, nil
}

// slice method.
func (s *SimpleBook) slice(filter func(*Entry) bool) []*Entry {
	entries := make([]*Entry, 0)
	for _, e := range s.entries {
		if !filter(e) {
			continue
		}
		entries = append(entries, e)
	}
	return entries
}

// byPriority helps us sort peer entries by priority.
type byPriority []*Entry

// Len returns the count of peer entries..
func (b byPriority) Len() int {
	return len(b)
}

// Less checks whether the score of the first peer is lower than the score of the second peer.
func (b byPriority) Less(i int, j int) bool {
	return b[i].score() > b[j].score()
}

// Swap will switch two peer entry positions in the list.
func (b byPriority) Swap(i int, j int) {
	b[i], b[j] = b[j], b[i]
}

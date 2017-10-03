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

package network

type Action uint8

type Event uint8

const (
	ActionAdd Action = iota
	ActionWhitelist
	ActionBlacklist
)

const (
	EventConnected Event = iota
	EventDisconnected
	EventDropped
	EventFailed
)

// Entry represents an entry in the simple address book, containing the address, whether the peer is
// currently active and how many successes/failures we had on the connection.
type Entry struct {
	Address    string
	Properties map[string]interface{}
}

// Float returns a float property for the entry.
func (e Entry) Float64(key string) float64 {
	val, ok := e.Properties[key]
	if !ok {
		return 0
	}
	floatVal, _ := val.(float64)
	return floatVal
}

// Int returns an integer property for the entry.
func (e Entry) Int(key string) int {
	val, ok := e.Properties[key]
	if !ok {
		return 0
	}
	intVal, _ := val.(int)
	return intVal
}

// Bool returns a bool property for the entry.
func (e Entry) Bool(key string) bool {
	val, ok := e.Properties[key]
	if !ok {
		return false
	}
	boolVal, _ := val.(bool)
	return boolVal
}

// Set sets a specific property to a given value.
func (e Entry) Set(key string, val interface{}) {
	e.Properties[key] = val
}

// Book represents an address book interface to handle known peer addresses on the alvalor network.
type Book interface {
	Action(address string, action Action)
	Event(address string, event Event)
	Sample(filter func(*Entry) bool, sort func([]*Entry) []*Entry, limit uint) ([]string, error)
}

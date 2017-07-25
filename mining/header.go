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

package mining

import (
	"encoding/binary"
	"time"

	"golang.org/x/crypto/blake2b"
)

// Header represents the header of a block.
type Header struct {
	Timestamp time.Time // timestamp when the block was mined
	State     []byte    // patricia merkle tree root hash of current state
	Delta     []byte    // patricia merkle tree root hash of transactions
	Previous  []byte    // hash / ID of the previous block
	Creator   []byte    // account of the creator of the block (for reward)
	Nonce     uint64    // nonce to modify the hash and comform to difficulty
}

// ID returns the ID of this block / header.
func (hdr Header) ID() []byte {
	h, _ := blake2b.New256(nil)
	ts, _ := hdr.Timestamp.MarshalBinary()
	h.Write(ts)
	h.Write(hdr.State)
	h.Write(hdr.Delta)
	h.Write(hdr.Previous)
	h.Write(hdr.Creator)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, hdr.Nonce)
	h.Write(buf)
	return h.Sum(nil)
}

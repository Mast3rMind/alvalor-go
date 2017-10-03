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
	Number     uint64    // block number / height of the block
	Nonce      uint64    // nonce to modify the hash and comform to difficulty
	Timestamp  time.Time // timestamp when the block was mined
	Previous   []byte    // hash / ID of the previous block
	Difficulty []byte    // target mining difficulty for block
	Creator    []byte    // account of the creator of the block (for reward)
	Data       []byte    // extra data that can be added to the block
	State      []byte    // patricia merkle tree root hash of account states
	Delta      []byte    // patricia merkle tree root hash of transactions
	Uncles     []byte    // patricia merkle tree root hash of uncles
}

// ID returns the ID of this block / header.
func (hdr Header) ID() []byte {
	h, _ := blake2b.New256(nil)

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, hdr.Number)
	h.Write(buf)
	binary.LittleEndian.PutUint64(buf, hdr.Nonce)
	h.Write(buf)
	ts, _ := hdr.Timestamp.MarshalBinary()
	h.Write(ts)

	h.Write(hdr.Previous)
	h.Write(hdr.Difficulty)
	h.Write(hdr.Creator)
	h.Write(hdr.Data)
	h.Write(hdr.State)
	h.Write(hdr.Delta)
	h.Write(hdr.Uncles)

	return h.Sum(nil)
}

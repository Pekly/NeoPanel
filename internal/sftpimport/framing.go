package sftpimport

import (
	"encoding/binary"
	"io"
	"sync/atomic"
)

const maxPacket = 256 << 10
const maxMetadata = 32 << 20
const maxEntries = 100000

// frameReader guards pkg/sftp v1.13.6's allocations and unsafe response
// unmarshalling. A length prefix is checked BEFORE allocation. Complete bounded
// frames are structurally checked before pkg/sftp sees them, including nested
// string/count/attribute lengths. This is not a second SFTP client.
// Metadata has a lifetime budget: even ReadDir (which buffers a whole directory)
// cannot accumulate unbounded entries. DATA has a separate lifetime byte budget.
type frameReader struct {
	source        io.Reader
	pending       []byte
	metadata      int64
	data          int64
	maxData       int64
	names         int
	listing       atomic.Bool
	readRemaining atomic.Int64
	failed        error
}

func (r *frameReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.failed != nil {
		return 0, r.failed
	}
	if len(r.pending) == 0 {
		var h [4]byte
		if _, err := io.ReadFull(r.source, h[:]); err != nil {
			r.failed = err
			return 0, err
		}
		n := binary.BigEndian.Uint32(h[:])
		if n < 5 || n > maxPacket {
			r.failed = problem("invalid SFTP packet size")
			return 0, r.failed
		}
		b := make([]byte, int(n)+4)
		copy(b, h[:])
		if _, err := io.ReadFull(r.source, b[4:]); err != nil {
			r.failed = err
			return 0, err
		}
		dataBytes, names, ok := validResponse(b[4:], r.listing.Load())
		if !ok {
			r.failed = problem("invalid SFTP response")
			return 0, r.failed
		}
		if b[4] == 103 {
			// pkg/sftp silently truncates an overlong DATA response to its caller's
			// buffer. Refuse that ambiguity, unsolicited data and zero-progress loops.
			if dataBytes == 0 || r.readRemaining.Add(-int64(dataBytes)) < 0 {
				r.failed = problem("invalid SFTP data length")
				return 0, r.failed
			}
			r.data += int64(dataBytes)
		} else {
			r.metadata += int64(n) + 4
		}
		r.names += names
		if r.metadata > maxMetadata || r.data > r.maxData || r.names > maxEntries+1 {
			r.failed = problem("SFTP response limit exceeded")
			return 0, r.failed
		}
		r.pending = b
	}
	n := copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

type cursor struct {
	b  []byte
	ok bool
}

func (c *cursor) take(n int) []byte {
	if !c.ok || n < 0 || n > len(c.b) {
		c.ok = false
		return nil
	}
	b := c.b[:n]
	c.b = c.b[n:]
	return b
}
func (c *cursor) u32() uint32 {
	b := c.take(4)
	if b == nil {
		return 0
	}
	return binary.BigEndian.Uint32(b)
}
func (c *cursor) str(max int) []byte {
	n := c.u32()
	if n > uint32(max) {
		c.ok = false
		return nil
	}
	return c.take(int(n))
}
func (c *cursor) attrs() {
	flags := c.u32()
	if flags & ^uint32(0x8000000f) != 0 {
		c.ok = false
		return
	}
	if flags&1 != 0 {
		c.take(8)
	}
	if flags&2 != 0 {
		c.take(8)
	}
	if flags&4 != 0 {
		c.take(4)
	}
	if flags&8 != 0 {
		c.take(8)
	}
	if flags&0x80000000 != 0 {
		n := c.u32()
		if n > uint32(len(c.b)/8) {
			c.ok = false
			return
		}
		for i := uint32(0); i < n && c.ok; i++ {
			c.str(maxPacket)
			c.str(maxPacket)
		}
	}
}
func validResponse(b []byte, listing bool) (dataBytes, names int, ok bool) {
	if len(b) < 5 {
		return 0, 0, false
	}
	c := cursor{b: b[1:], ok: true}
	c.u32() // request ID or VERSION number
	switch b[0] {
	case 2: // VERSION; this client speaks SFTP v3 only.
		if binary.BigEndian.Uint32(b[1:5]) != 3 {
			return 0, 0, false
		}
		for len(c.b) > 0 && c.ok {
			c.str(4096)
			c.str(4096)
		}
	case 101: // STATUS
		c.u32()
		c.str(maxPacket)
		c.str(4096)
	case 102: // HANDLE
		c.str(4096)
	case 103: // DATA
		dataBytes = len(c.str(32768))
	case 104: // NAME, from REALPATH or READDIR
		n := c.u32()
		if n > uint32(len(c.b)/12) || n > maxEntries {
			return 0, 0, false
		}
		names = int(n)
		for i := uint32(0); i < n && c.ok; i++ {
			name := string(c.str(4096))
			c.str(16384)
			c.attrs()
			if listing {
				if name != "." && name != ".." && !safeComponent(name) {
					c.ok = false
				}
			} else if !safePath(name, true) {
				c.ok = false
			}
		}
	case 105:
		c.attrs() // ATTRS
	default:
		return 0, 0, false
	}
	return dataBytes, names, c.ok && len(c.b) == 0
}

package chunker

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

// ValidateLevel parses a bounded, structurally valid NBT document and checks the
// exact writer version, not just gzip magic or an arbitrary embedded byte string.
func ValidateLevel(root *rooted.Root, f Format) error {
	file, err := root.Open("level.dat")
	if err != nil {
		return err
	}
	defer file.Close()
	raw, err := rooted.ReadBounded(file, 16<<20)
	if err != nil {
		return err
	}
	var order binary.ByteOrder = binary.BigEndian
	if f.Edition == "java" {
		gz, e := gzip.NewReader(bytes.NewReader(raw))
		if e != nil {
			return e
		}
		raw, err = rooted.ReadBounded(gz, 16<<20)
		gz.Close()
		if err != nil {
			return err
		}
	} else {
		if len(raw) < 8 || (binary.LittleEndian.Uint32(raw[:4]) != 8 && binary.LittleEndian.Uint32(raw[:4]) != 10) || int(binary.LittleEndian.Uint32(raw[4:8])) != len(raw)-8 {
			return errors.New("invalid Bedrock level.dat header")
		}
		raw = raw[8:]
		order = binary.LittleEndian
	}
	p := nbtParser{r: bytes.NewReader(raw), order: order, ints: map[string]int32{}}
	tag, e := p.r.ReadByte()
	if e != nil || tag != 10 {
		return errors.New("level.dat must have an NBT compound root")
	}
	if _, e = p.text(); e != nil {
		return e
	}
	if e = p.payload(10, "", 0); e != nil {
		return fmt.Errorf("invalid level.dat: %w", e)
	}
	if p.r.Len() != 0 {
		return errors.New("trailing level.dat data")
	}
	if f.Edition == "bedrock" {
		parts := strings.Split(f.Version, ".")
		if len(p.bedrockVersion) < 3 {
			return errors.New("missing Bedrock lastOpenedWithVersion")
		}
		for i := 0; i < 3; i++ {
			want := 0
			if i < len(parts) {
				want, _ = strconv.Atoi(parts[i])
			}
			if p.bedrockVersion[i] != int32(want) {
				return errors.New("Bedrock output version mismatch")
			}
		}
	} else if got, ok := p.ints["Data/DataVersion"]; !ok || got != f.DataVersion {
		return errors.New("converted level.dat does not match requested edition/version")
	}
	return nil
}

type nbtParser struct {
	r              *bytes.Reader
	order          binary.ByteOrder
	nodes          int
	ints           map[string]int32
	bedrockVersion []int32
}

func (p *nbtParser) text() (string, error) {
	var n uint16
	if e := binary.Read(p.r, p.order, &n); e != nil {
		return "", e
	}
	b := make([]byte, int(n))
	_, e := io.ReadFull(p.r, b)
	return string(b), e
}
func (p *nbtParser) skip(n int64) error {
	if n < 0 || n > int64(p.r.Len()) {
		return io.ErrUnexpectedEOF
	}
	_, e := p.r.Seek(n, io.SeekCurrent)
	return e
}
func (p *nbtParser) length() (int32, error) {
	var n int32
	e := binary.Read(p.r, p.order, &n)
	if e == nil && (n < 0 || n > 1<<20) {
		e = errors.New("NBT collection limit exceeded")
	}
	return n, e
}
func (p *nbtParser) payload(t byte, path string, depth int) error {
	p.nodes++
	if depth > 64 || p.nodes > 1<<20 {
		return errors.New("NBT complexity limit exceeded")
	}
	switch t {
	case 1:
		return p.skip(1)
	case 2:
		return p.skip(2)
	case 3:
		var v int32
		if e := binary.Read(p.r, p.order, &v); e != nil {
			return e
		}
		if (path == "Data/DataVersion" && depth == 2) || (path == "NetworkVersion" && depth == 1) {
			if _, ok := p.ints[path]; ok {
				return errors.New("duplicate version tag")
			}
			p.ints[path] = v
		}
		return nil
	case 4, 6:
		return p.skip(8)
	case 5:
		return p.skip(4)
	case 7, 11, 12:
		n, e := p.length()
		if e != nil {
			return e
		}
		unit := int64(1)
		if t == 11 {
			unit = 4
		}
		if t == 12 {
			unit = 8
		}
		return p.skip(int64(n) * unit)
	case 8:
		_, e := p.text()
		return e
	case 9:
		child, e := p.r.ReadByte()
		if e != nil {
			return e
		}
		n, e := p.length()
		if e != nil {
			return e
		}
		if path == "lastOpenedWithVersion" && depth == 1 {
			if child != 3 || n < 3 || n > 8 || p.bedrockVersion != nil {
				return errors.New("invalid Bedrock version list")
			}
			p.bedrockVersion = make([]int32, n)
			return binary.Read(p.r, p.order, p.bedrockVersion)
		}
		for i := int32(0); i < n; i++ {
			if e = p.payload(child, path+"[]", depth+1); e != nil {
				return e
			}
		}
		return nil
	case 10:
		for {
			child, e := p.r.ReadByte()
			if e != nil {
				return e
			}
			if child == 0 {
				return nil
			}
			name, e := p.text()
			if e != nil {
				return e
			}
			next := name
			if path != "" {
				next = path + "/" + name
			}
			if e = p.payload(child, next, depth+1); e != nil {
				return e
			}
		}
	default:
		return errors.New("unsupported NBT tag")
	}
}

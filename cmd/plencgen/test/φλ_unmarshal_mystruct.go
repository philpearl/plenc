package test

import (
	"fmt"
	"time"

	"github.com/philpearl/plenc"
)

var _ time.Time

func (e *MyStruct) ΦλUnmarshal(data []byte) (int, error) {

	var offset int
	for offset < len(data) {
		wt, index, n := plenc.ReadTag(data[offset:])
		if n == 0 {
			break
		}
		offset += n
		switch index {

		case 1:

			// intdecode
			v, n := plenc.ReadVarInt(data[offset:])
			e.A = int(v)

			offset += n

		case 2:

			v, n := plenc.ReadVarUint(data[offset:])
			e.B = uint(v)

			offset += n

		case 3:

			v, n := plenc.ReadVarUint(data[offset:])
			e.C = uint32(v)

			offset += n

		case 4:

			v, n := plenc.ReadFloat32(data[offset:])
			e.D = float32(v)

			offset += n

		case 5:

			v, n := plenc.ReadFloat64(data[offset:])
			e.E = float64(v)

			offset += n

		case 6:

			v, n := plenc.ReadBool(data[offset:])
			e.F = bool(v)

			offset += n

		case 7:

			s, n := plenc.ReadVarUint(data[offset:])
			offset += n
			n, err := e.H.ΦλUnmarshal(data[offset : offset+int(s)])
			if err != nil {
				return 0, fmt.Errorf("failed to unmarshal field %d H (Struct2). %w", index, err)
			}

			offset += n

		case 8:

			// Slice of method-y things. Good to grow the slice first in case it is large
			l := len(e.I)
			if cap(e.I) > l {
				e.I = e.I[:l+1]
			} else {
				// Need to grow slice. What's the best way?!
				e.I = append(e.I, Struct2{})
			}

			// Slice of method-y things
			s, n := plenc.ReadVarUint(data[offset:])
			offset += n
			n, err := e.I[l].ΦλUnmarshal(data[offset : offset+int(s)])
			if err != nil {
				return 0, fmt.Errorf("failed to unmarshal field %d I (Struct2). %w", index, err)
			}

			offset += n

		case 9:

			if e.J == nil {
				e.J = new(Struct2)
			}

			s, n := plenc.ReadVarUint(data[offset:])
			offset += n
			n, err := e.J.ΦλUnmarshal(data[offset : offset+int(s)])
			if err != nil {
				return 0, fmt.Errorf("failed to unmarshal field %d J (Struct2). %w", index, err)
			}

			offset += n

		case 10:

			// Slice of method-y things. Good to grow the slice first in case it is large
			l := len(e.K)
			e.K = append(e.K, &Struct2{})

			// Slice of method-y things
			s, n := plenc.ReadVarUint(data[offset:])
			offset += n
			n, err := e.K[l].ΦλUnmarshal(data[offset : offset+int(s)])
			if err != nil {
				return 0, fmt.Errorf("failed to unmarshal field %d K (Struct2). %w", index, err)
			}

			offset += n

		case 11:

			// intdecode
			v, n := plenc.ReadVarInt(data[offset:])
			e.L = FunnyInt(v)

			offset += n

		case 12:

			{
				var (
					t   plenc.Time
					s   uint64
					err error
				)
				s, n = plenc.ReadVarUint(data[offset:])
				offset += n
				n, err = t.ΦλUnmarshal(data[offset : offset+int(s)])
				if err != nil {
					return 0, fmt.Errorf("failed to unmarshal field %d M (time.Time). %w", index, err)
				}
				e.M = t.Standard()
			}

			offset += n

		case 13:

			if e.N == nil {
				e.N = new(time.Time)
			}

			{
				var (
					t   plenc.Time
					s   uint64
					err error
				)
				s, n = plenc.ReadVarUint(data[offset:])
				offset += n
				n, err = t.ΦλUnmarshal(data[offset : offset+int(s)])
				if err != nil {
					return 0, fmt.Errorf("failed to unmarshal field %d N (time.Time). %w", index, err)
				}
				*e.N = t.Standard()
			}

			offset += n

		case 16:

			n, err := e.Q.ΦλUnmarshal(data[offset:])
			if err != nil {
				return 0, fmt.Errorf("failed to unmarshal field %d Q (OptInt). %w", index, err)
			}

			offset += n

		default:
			// Field corresponding to index does not exist
			n, err := plenc.Skip(data[offset:], wt)
			if err != nil {
				return 0, fmt.Errorf("failed to skip field %d. %w", index, err)
			}
			offset += n
		}
	}

	return offset, nil
}

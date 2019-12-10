package example

import (
	"fmt"
	"time"

	"github.com/philpearl/plenc"
)

var _ time.Time

func (e *example) ΦλUnmarshal(data []byte) (int, error) {

	var offset int
	for offset < len(data) {
		wt, index, n := plenc.ReadTag(data[offset:])
		if n == 0 {
			break
		}
		offset += n
		switch index {

		case 1:

			v, n := plenc.ReadString(data[offset:])
			e.Name = string(v)

			offset += n

		case 2:

			// intdecode
			v, n := plenc.ReadVarInt(data[offset:])
			e.Age = int(v)

			offset += n

		case 3:

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
					return 0, fmt.Errorf("failed to unmarshal field %d Starting (time.Time). %w", index, err)
				}
				e.Starting = t.Standard()
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

package test

import (
	"fmt"
	"time"

	"github.com/philpearl/φλenc"
)

var _ time.Time

func (e *Struct2) ΦλUnmarshal(data []byte) (int, error) {

	var offset int
	for offset < len(data) {
		wt, index, n := φλenc.ReadTag(data[offset:])
		if n == 0 {
			break
		}
		offset += n
		switch index {

		case 1:

			v, n := φλenc.ReadVarUint(data[offset:])
			e.A = uint8(v)

			offset += n

		case 2:

			v, n := φλenc.ReadString(data[offset:])
			e.B = string(v)

			offset += n

		default:
			// Field corresponding to index does not exist
			n, err := φλenc.Skip(data[offset:], wt)
			if err != nil {
				return 0, fmt.Errorf("failed to skip field %d. %w", index, err)
			}
			offset += n
		}
	}

	return offset, nil
}

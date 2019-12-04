package philenc

import "testing"

func TestTags(t *testing.T) {
	for wt := WTVarInt; wt <= WT32; wt++ {
		for index := 1; index < 1000; index++ {
			l := SizeTag(wt, index)
			b := make([]byte, 0, l)
			data := AppendTag(b, wt, index)
			awt, aindex, n := ReadTag(data)

			if n != l {
				t.Errorf("data size issue %d %d", l, n)
			}
			if awt != wt || aindex != index {
				t.Errorf("exp %d %d, actual %d %d", wt, index, awt, aindex)
			}
		}
	}
}

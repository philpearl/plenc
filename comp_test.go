package φλenc

import (
	fmt "fmt"
	"testing"

	proto "github.com/golang/protobuf/proto"
)

func TestCompProto(t *testing.T) {
	v := Trial{
		F1: 42,
		F2: []int32{6, 7, 200},
		F3: 3.14,
		F4: []float32{3.14, 7, 2992},
	}

	data, err := proto.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%X\n", data)
}

/*
082A
12040607C801
1D 14AE4740
220C C3F54840 0000E04000003B45

082A
12040607C801
1D C3F54840
22 0C C3F54840 0000E040 00003B45

[8 42 18 3 6 7 8 29 20 174 71 64 34 12 195 245 72 64 0 0 224 64 0 0 59 69]

[8 42
18 4 6 7 200 1
29 20 174 71 64
34 12 195 245 72
64 0 0 224 64 0 0 59 69]

*/

package philenc

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
)

type Cheese struct {
	Fred int64 `protobuf:"varint,1,req,name=fred,proto3"`
	Jim []int64 `protobuf:"varint,2,req,name=jim,proto3"`
	Sheila []float32 `protobuf:"float32,3,req,name=jim,proto3"`
}

func (c *Cheese) Reset() {
	c.Fred = 0
}

func (c *Cheese) String() string { return "Cheese" }
func (c *Cheese) ProtoMessage()  {}

type CustomerProfileRequest struct {
	CustomerId string `protobuf:"bytes,1,opt,name=customerId,proto3" json:"customerId,omitempty"`
	La         Cheese `protobuf:"bytes,2,req,name=la,proto3"`
}

func (c *CustomerProfileRequest) Reset() {
	c.CustomerId = ""
}

func (c *CustomerProfileRequest) String() string {
	return "CustomerProfileRequest"
}

func (c *CustomerProfileRequest) ProtoMessage() {}

func TestBasicProto(t *testing.T) {
	var v CustomerProfileRequest

	v.CustomerId = "blah"
	v.La.Fred = 0x42

	{
		data, err := proto.Marshal(&v.La)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(data)

	}

	{
		data, err := proto.Marshal(&v)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(data)

		var w CustomerProfileRequest
		if err := proto.Unmarshal(data, &w); err != nil {
			t.Fatal(err)
		}
		fmt.Println(w)
	}
}

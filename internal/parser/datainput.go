package parser

import (
	"encoding/binary"
	"fmt"
	"math"
)

type DataInput struct {
	Data         []byte
	IndexPointer int
	Length       int
}

func NewDataInput(data []byte) *DataInput {
	return &DataInput{
		Data:         data,
		IndexPointer: 0,
		Length:       len(data),
	}
}

func (di *DataInput) Read(n int) []byte {
	if di.IndexPointer+n > len(di.Data) {
		return nil
	}
	data := di.Data[di.IndexPointer : di.IndexPointer+n]
	di.IndexPointer += n
	return data
}

func (di *DataInput) ReadInt() int {
	return int(int32(binary.BigEndian.Uint32(di.Read(4))))
}

func (di *DataInput) ReadUint() uint {
	return uint(binary.BigEndian.Uint32(di.Read(4)))
}

func (di *DataInput) ReadUnsignedShort() uint16 {
	return binary.BigEndian.Uint16(di.Read(2))
}

func (di *DataInput) ReadUTF() string {
	lon := int(di.ReadUnsignedShort())
	return string(di.Read(lon))
}

func (di *DataInput) ReadBoolean() bool {
	ans := di.Read(1)
	return ans[0] == 1
}

func (di *DataInput) ReadDouble() float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(di.Read(8)))
}

func (di *DataInput) ReadUnsignedByte() uint8 {
	return di.Read(1)[0]
}

func (di *DataInput) ReadVarInt() int {
	ans := 0
	for i := 0; i < 32; i += 7 {
		b := di.ReadUnsignedByte()
		ans |= int(b&0b01111111) << i
		if b&0b10000000 == 0 {
			return ans
		}
	}
	panic("Too much data")
}

func (di *DataInput) ReadVarUhInt() int {
	return di.ReadVarInt()
}

func (di *DataInput) AreBytesAvailable() bool {
	return di.IndexPointer < len(di.Data)
}

func (di *DataInput) SetPointer(pointer int) {
	di.IndexPointer = pointer
}

func (di *DataInput) OffsetStr() string {
	return fmt.Sprintf("%#x (%d)", di.IndexPointer, di.IndexPointer)
}

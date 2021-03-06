package bin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testBuffer = []byte("1234567890abcdefghijklmnopqrstuvwxyz")

/*
===============================================================================
    Reader
===============================================================================
*/

func TestNewReader(t *testing.T) {
	t.Parallel()
	bb := NewReaderBytes(testBuffer, binary.LittleEndian)
	assert.Equal(t, int64(0), bb.GetPosition())
}

func TestReadByte(t *testing.T) {
	t.Parallel()
	bb := NewReaderBytes(testBuffer, binary.LittleEndian)
	c := byte(0)
	for _, expected := range testBuffer {
		err := bb.ReadByte(&c)
		assert.NoError(t, err)
		assert.Equal(t, expected, c)
	}
}
func TestReadByteError(t *testing.T) {
	t.Parallel()
	bb := NewReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF
	bb.source.Read(make([]byte, len(testBuffer)))
	c := byte(0)
	err := bb.ReadByte(&c)
	assert.Error(t, err)
}

func TestRead(t *testing.T) {
	t.Parallel()
	bb := NewReaderBytes(testBuffer, binary.LittleEndian)
	buf := make([]byte, 32)
	nread, err := bb.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), nread)
	assert.Equal(t, testBuffer[:32], buf[:32])
}

func TestReadBytes(t *testing.T) {
	t.Parallel()
	buf = append(testBuffer, make([]byte, 1024)...)
	bb := NewReaderBytes(buf, binary.LittleEndian)

	tmp := make([]byte, 2048)

	err := bb.ReadBytes(tmp[:2])
	assert.NoError(t, err)
	assert.Equal(t, buf[0:2], tmp[0:2])

	err = bb.ReadBytes(tmp[:3])
	assert.NoError(t, err)
	assert.Equal(t, buf[2:5], tmp[:3])

	err = bb.ReadBytes(tmp[:1])
	assert.NoError(t, err)
	assert.Equal(t, byte('6'), tmp[0])

	// big read should alloc new buffer
	err = bb.ReadBytes(tmp[:1028])
	assert.NoError(t, err)
	assert.Equal(t, buf[6:1034], tmp[:1028])
}

func TestReadBytesError(t *testing.T) {
	t.Parallel()
	buf := make([]byte, 32)

	// nil reader
	bb := Reader{}
	err := bb.ReadBytes(buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadBytes(buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-10))
	err = bb.ReadBytes(buf)
	assert.Error(t, err)
}

func TestReadUint16(t *testing.T) {
	t.Parallel()
	// Little Endian
	buf := []byte{0x08, 0x00, 0xFF, 0x01}
	bb := NewReaderBytes(buf, binary.LittleEndian)
	ui16 := uint16(0)

	err := bb.ReadUint16(&ui16)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0x0008), ui16)

	err = bb.ReadUint16(&ui16)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0x01FF), ui16)

	// Big Endian
	buf = []byte{0x08, 0x00, 0xFF, 0x01}
	bb = NewReaderBytes(buf, binary.BigEndian)

	err = bb.ReadUint16(&ui16)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0x0800), ui16)

	err = bb.ReadUint16(&ui16)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0xFF01), ui16)
}

func TestReadUint16Error(t *testing.T) {
	t.Parallel()
	buf := uint16(0)

	// nil reader
	bb := Reader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadUint16(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = Reader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadUint16(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadUint16(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-1))
	err = bb.ReadUint16(&buf)
	assert.Error(t, err)
}

func TestReadUint32(t *testing.T) {
	t.Parallel()
	// Little Endian
	buf := []byte{0x08, 0x00, 0xFF, 0x01}
	bb := NewReaderBytes(buf, binary.LittleEndian)
	ui32 := uint32(0)

	err := bb.ReadUint32(&ui32)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0x01FF0008), ui32)

	// Big Endian
	buf = []byte{0x08, 0x00, 0xFF, 0x01}
	bb = NewReaderBytes(buf, binary.BigEndian)
	err = bb.ReadUint32(&ui32)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0x0800FF01), ui32)
}

func TestReadUint32Error(t *testing.T) {
	t.Parallel()
	buf := uint32(0)

	// nil reader
	bb := Reader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadUint32(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = Reader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadUint32(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadUint32(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-3))
	err = bb.ReadUint32(&buf)
	assert.Error(t, err)
}

func TestReadUint64(t *testing.T) {
	t.Parallel()
	// Little Endian
	buf := []byte{0x08, 0x00, 0xFF, 0x01, 0x08, 0x00, 0xFF, 0x01}
	bb := NewReaderBytes(buf, binary.LittleEndian)
	ui64 := uint64(0)

	err := bb.ReadUint64(&ui64)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x01FF000801FF0008), ui64)

	// Big Endian
	buf = []byte{0x08, 0x00, 0xFF, 0x01, 0x08, 0x00, 0xFF, 0x01}
	bb = NewReaderBytes(buf, binary.BigEndian)
	err = bb.ReadUint64(&ui64)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x800FF010800FF01), ui64)
}

func TestReadUint64Error(t *testing.T) {
	t.Parallel()
	buf := uint64(0)

	// nil reader
	bb := Reader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadUint64(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = Reader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadUint64(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadUint64(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-3))
	err = bb.ReadUint64(&buf)
	assert.Error(t, err)
}

func TestReadFloat32(t *testing.T) {
	t.Parallel()
	f32 := float32(0)

	// Little Endian
	buf := []byte{0x79, 0xe9, 0xf6, 0x42}
	bb := NewReaderBytes(buf, binary.LittleEndian)
	err := bb.ReadFloat32(&f32)
	assert.NoError(t, err)
	assert.Equal(t, float32(123.456), f32)

	// Big Endian
	buf = []byte{0x42, 0xf6, 0xe9, 0x79}
	bb = NewReaderBytes(buf, binary.BigEndian)
	err = bb.ReadFloat32(&f32)
	assert.NoError(t, err)
	assert.Equal(t, float32(123.456), f32)
}

func TestReadFloat32Error(t *testing.T) {
	t.Parallel()
	buf := float32(0)

	// nil reader
	bb := Reader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadFloat32(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = Reader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadFloat32(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadFloat32(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-2))
	err = bb.ReadFloat32(&buf)
	assert.Error(t, err)
}

func TestReadFloat64(t *testing.T) {
	t.Parallel()
	f64 := float64(0)

	// Little Endian
	buf := []byte{0x77, 0xBE, 0x9F, 0x1A, 0x2F, 0xDD, 0x5E, 0x40}
	bb := NewReaderBytes(buf, binary.LittleEndian)
	err := bb.ReadFloat64(&f64)
	assert.NoError(t, err)
	assert.Equal(t, float64(123.456), f64)

	// Big Endian
	buf = []byte{0x40, 0x5E, 0xDD, 0x2F, 0x1A, 0x9F, 0xBE, 0x77}
	bb = NewReaderBytes(buf, binary.BigEndian)
	err = bb.ReadFloat64(&f64)
	assert.NoError(t, err)
	assert.Equal(t, float64(123.456), f64)
}

func TestReadFloat64Error(t *testing.T) {
	t.Parallel()
	buf := float64(0)

	// nil reader
	bb := Reader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadFloat64(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = Reader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadFloat64(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadFloat64(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-1))
	err = bb.ReadFloat64(&buf)
	assert.Error(t, err)
}

func TestPeek(t *testing.T) {
	t.Parallel()
	bb := NewReaderBytes(testBuffer, binary.LittleEndian)
	// peek +3
	pb := make([]byte, 3)
	assert.NoError(t, bb.Peek(pb))
	assert.Equal(t, []byte("123"), pb)
	// should still be at position zero
	assert.Equal(t, int64(0), bb.GetPosition())

	// read +2, should equal "12"
	buf := make([]byte, 2)
	assert.NoError(t, bb.ReadBytes(buf))
	assert.Equal(t, []byte("12"), buf)
	assert.Equal(t, int64(2), bb.GetPosition())

	// peek +1
	pb2 := make([]byte, 1)
	assert.NoError(t, bb.Peek(pb2))
	assert.Equal(t, []byte("3"), pb2)
	// should be at position two
	assert.Equal(t, int64(2), bb.GetPosition())

	// read +4, should equal "3456"
	buf = make([]byte, 4)
	assert.NoError(t, bb.ReadBytes(buf))
	assert.Equal(t, []byte("3456"), buf)
	// should be at position 6
	assert.Equal(t, int64(6), bb.GetPosition())

	// peek +0
	assert.NoError(t, bb.Peek([]byte{}))
	// should be at position 6
	assert.Equal(t, int64(6), bb.GetPosition())
}

func TestMultiplePeek(t *testing.T) {
	t.Parallel()
	bb := NewReaderBytes(testBuffer, binary.LittleEndian)
	// peek +3
	pb := make([]byte, 3)
	assert.NoError(t, bb.Peek(pb))
	assert.Equal(t, []byte("123"), pb)
	// should still be at position zero
	assert.Equal(t, int64(0), bb.GetPosition())

	// peek +1
	pb2 := make([]byte, 1)
	assert.NoError(t, bb.Peek(pb2))
	assert.Equal(t, []byte("1"), pb2)
	// should still be at position zero
	assert.Equal(t, int64(0), bb.GetPosition())

	// read +2, should equal "12"
	buf := make([]byte, 2)
	assert.NoError(t, bb.ReadBytes(buf))
	assert.Equal(t, []byte("12"), buf)
	assert.Equal(t, int64(2), bb.GetPosition())

	// read +4, should equal "3456"
	buf = make([]byte, 4)
	assert.NoError(t, bb.ReadBytes(buf))
	assert.Equal(t, []byte("3456"), buf)
	assert.Equal(t, int64(6), bb.GetPosition())
}

func TestPeakLarge(t *testing.T) {
	t.Parallel()
	bb := NewReader(blackHole, binary.LittleEndian)

	// peek +80
	pb := make([]byte, 80)
	assert.NoError(t, bb.Peek(pb))
	assert.Equal(t, int64(0), bb.GetPosition())

	// peek +8192
	pb = make([]byte, 8192)
	assert.NoError(t, bb.Peek(pb))
	assert.Equal(t, int64(0), bb.GetPosition())

	// read +1024, should equal 0x00 x 1024
	buf := make([]byte, 1024)
	assert.NoError(t, bb.ReadBytes(buf))
	assert.Equal(t, make([]byte, 1024), buf)
	assert.Equal(t, int64(1024), bb.GetPosition())

	bb = NewReader(blackHole, binary.LittleEndian)

	// peek +2048
	pb = make([]byte, 2048)
	assert.NoError(t, bb.Peek(pb))
	assert.Equal(t, int64(0), bb.GetPosition())

	// read +8192, should equal 0x00 x 8192
	buf = make([]byte, 8192)
	assert.NoError(t, bb.ReadBytes(buf))
	assert.Equal(t, make([]byte, 8192), buf)
	assert.Equal(t, int64(8192), bb.GetPosition())

}

func TestPeekError(t *testing.T) {
	t.Parallel()
	buf := make([]byte, 4)

	// nil reader
	bb := Reader{}
	bb.bo = binary.LittleEndian
	assert.Error(t, bb.Peek(buf))

	// Reached EOF
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	assert.Error(t, bb.Peek(buf))

	// Reached EOF during read (partial)
	bb = NewReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-1))
	assert.Error(t, bb.Peek(buf))
}

func TestDiscard(t *testing.T) {
	t.Parallel()
	bb := NewReader(blackHole, binary.LittleEndian)
	bb.Discard(0)
	assert.Equal(t, int64(0), bb.GetPosition())
	bb.Discard(1)
	assert.Equal(t, int64(1), bb.GetPosition())
	bb.Discard(7)
	assert.Equal(t, int64(8), bb.GetPosition())
	bb.Discard(2)
	assert.Equal(t, int64(10), bb.GetPosition())

	// discard (relatively) large amount of bytes
	expectedPos := int64(10)
	for _, v := range []int64{800, 1024, 1080, 4096, 8000} {
		bb.Discard(v)
		expectedPos += v
		assert.Equal(t, expectedPos, bb.GetPosition())
	}

}

func TestDiscardError(t *testing.T) {
	t.Parallel()
	// nil reader
	bb := Reader{}
	bb.bo = binary.LittleEndian
	assert.Error(t, bb.Discard(2))

	// reader error
	bb = NewReader(errRW, binary.LittleEndian)
	err := bb.Discard(2)
	assert.Error(t, err)

	// reader error with large enough discard to cause chunking
	err = bb.Discard(4096)
	assert.Error(t, err)
}

func TestReaderReset(t *testing.T) {
	t.Parallel()
	// Big Endian reader at position 10
	bb := NewReaderBytes(testBuffer, binary.BigEndian)
	assert.NoError(t, bb.Discard(10))
	assert.Equal(t, int64(10), bb.GetPosition())

	// Reset to Little Endian
	buf2 := []byte("9876543210")
	bb.Reset(bytes.NewReader(buf2), binary.LittleEndian)
	assert.Equal(t, int64(0), bb.GetPosition())
	assert.Equal(t, binary.LittleEndian, bb.GetByteOrder())
	assert.NoError(t, bb.Discard(4))
	assert.Equal(t, int64(4), bb.GetPosition())
}

/*
===============================================================================
    Writer
===============================================================================
*/

func TestNewWriter(t *testing.T) {
	t.Parallel()
	w := bytes.NewBuffer(testBuffer)
	bw := NewWriter(w, binary.LittleEndian)
	assert.Equal(t, int64(0), bw.GetPosition())
}

func TestWriteByte(t *testing.T) {
	t.Parallel()
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.LittleEndian)
	// put all of `testBuffer` into bw byte-by-byte
	for _, c := range testBuffer {
		err := bw.WriteByte(c)
		assert.NoError(t, err)
	}
	assert.Equal(t, int64(len(testBuffer)), bw.GetPosition())
	assert.Equal(t, testBuffer, w.Bytes())
}

func TestWriteByteError(t *testing.T) {
	t.Parallel()
	// nil writer
	bw := Writer{}
	bw.bo = binary.LittleEndian
	assert.Error(t, bw.WriteByte(0xFF))

	// writer error
	bw = NewWriter(errRW, binary.LittleEndian)
	assert.Error(t, bw.WriteByte(0xFF))
}

func TestWrite(t *testing.T) {
	t.Parallel()
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.LittleEndian)
	// put all of `testBuffer` into bw
	nwrite, err := bw.Write(testBuffer)
	assert.NoError(t, err)
	assert.Equal(t, len(testBuffer), nwrite)
	assert.Equal(t, int64(len(testBuffer)), bw.GetPosition())
	assert.Equal(t, testBuffer, w.Bytes())

	nwrite, err = bw.Write([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, 0, nwrite)
	assert.Equal(t, int64(len(testBuffer)), bw.GetPosition())
	assert.Equal(t, testBuffer, w.Bytes())
}

func TestWriteBytes(t *testing.T) {
	t.Parallel()
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.LittleEndian)
	// put all of `testBuffer` into bw
	err := bw.WriteBytes(testBuffer)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(testBuffer)), bw.GetPosition())
	assert.Equal(t, testBuffer, w.Bytes())

	assert.NoError(t, bw.WriteBytes([]byte{}))
}

func TestWriteBytesError(t *testing.T) {
	t.Parallel()
	// nil writer
	bw := Writer{}
	bw.bo = binary.LittleEndian
	assert.Error(t, bw.WriteBytes([]byte{0xFF}))

	// writer error
	bw = NewWriter(errRW, binary.LittleEndian)
	assert.Error(t, bw.WriteBytes([]byte{0xFF}))
}

func TestWriteUint16(t *testing.T) {
	t.Parallel()
	// Little Endian
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.LittleEndian)

	err := bw.WriteUint16(uint16(1234))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xD2, 0x04}, w.Bytes())
	assert.Equal(t, int64(2), bw.GetPosition())

	// Big Endian
	w = bytes.NewBuffer([]byte{})
	bw = NewWriter(w, binary.BigEndian)

	err = bw.WriteUint16(uint16(1234))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x04, 0xD2}, w.Bytes())
	assert.Equal(t, int64(2), bw.GetPosition())
}

func TestWriteUint16Error(t *testing.T) {
	t.Parallel()
	// nil writer
	bw := Writer{}
	bw.bo = binary.LittleEndian
	assert.Error(t, bw.WriteUint16(uint16(1234)))

	// nil byte order
	bw = Writer{}
	bw.dest = bytes.NewBuffer([]byte{})
	assert.Error(t, bw.WriteUint16(uint16(1234)))

	// writer error
	bw = NewWriter(errRW, binary.LittleEndian)
	assert.Error(t, bw.WriteUint16(uint16(1234)))
}

func TestWriteUint32(t *testing.T) {
	t.Parallel()
	// Little Endian
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.LittleEndian)

	err := bw.WriteUint32(uint32(1234))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xD2, 0x04, 0x00, 0x00}, w.Bytes())
	assert.Equal(t, int64(4), bw.GetPosition())

	// Big Endian
	w = bytes.NewBuffer([]byte{})
	bw = NewWriter(w, binary.BigEndian)

	err = bw.WriteUint32(uint32(1234))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x00, 0x04, 0xD2}, w.Bytes())
	assert.Equal(t, int64(4), bw.GetPosition())
}

func TestWriteUint32Error(t *testing.T) {
	t.Parallel()
	// nil writer
	bw := Writer{}
	bw.bo = binary.LittleEndian
	assert.Error(t, bw.WriteUint32(uint32(1234)))

	// nil byte order
	bw = Writer{}
	bw.dest = bytes.NewBuffer([]byte{})
	assert.Error(t, bw.WriteUint32(uint32(1234)))

	// writer error
	bw = NewWriter(errRW, binary.LittleEndian)
	assert.Error(t, bw.WriteUint32(uint32(1234)))
}

func TestWriteUint64(t *testing.T) {
	t.Parallel()
	// Little Endian
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.LittleEndian)

	err := bw.WriteUint64(uint64(123456789))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x15, 0xCD, 0x5B, 0x07, 0x00, 0x00, 0x00, 0x00}, w.Bytes())
	assert.Equal(t, int64(8), bw.GetPosition())

	// Big Endian
	w = bytes.NewBuffer([]byte{})
	bw = NewWriter(w, binary.BigEndian)

	err = bw.WriteUint64(uint64(123456789))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x07, 0x5B, 0xCD, 0x15}, w.Bytes())
	assert.Equal(t, int64(8), bw.GetPosition())
}

func TestWriteUint64Error(t *testing.T) {
	t.Parallel()
	// nil writer
	bw := Writer{}
	bw.bo = binary.LittleEndian
	assert.Error(t, bw.WriteUint64(uint64(1234)))

	// nil byte order
	bw = Writer{}
	bw.dest = bytes.NewBuffer([]byte{})
	assert.Error(t, bw.WriteUint64(uint64(1234)))

	// writer error
	bw = NewWriter(errRW, binary.LittleEndian)
	assert.Error(t, bw.WriteUint64(uint64(1234)))
}

func TestWriteFloat32(t *testing.T) {
	t.Parallel()
	// Little Endian
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.LittleEndian)

	err := bw.WriteFloat32(float32(1234.5678))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x2B, 0x52, 0x9A, 0x44}, w.Bytes())
	assert.Equal(t, int64(4), bw.GetPosition())

	// Big Endian
	w = bytes.NewBuffer([]byte{})
	bw = NewWriter(w, binary.BigEndian)

	err = bw.WriteFloat32(float32(1234.5678))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x44, 0x9A, 0x52, 0x2B}, w.Bytes())
	assert.Equal(t, int64(4), bw.GetPosition())
}

func TestWriteFloat32Error(t *testing.T) {
	t.Parallel()
	// nil writer
	bw := Writer{}
	bw.bo = binary.LittleEndian
	assert.Error(t, bw.WriteFloat32(float32(1234.5678)))

	// nil byte order
	bw = Writer{}
	bw.dest = bytes.NewBuffer([]byte{})
	assert.Error(t, bw.WriteFloat32(float32(1234.5678)))

	// writer error
	bw = NewWriter(errRW, binary.LittleEndian)
	assert.Error(t, bw.WriteFloat32(float32(1234.5678)))
}

func TestWriteFloat64(t *testing.T) {
	t.Parallel()
	// Little Endian
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.LittleEndian)

	err := bw.WriteFloat64(float64(1234.5678))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xAD, 0xFA, 0x5C, 0x6D, 0x45, 0x4A, 0x93, 0x40}, w.Bytes())
	assert.Equal(t, int64(8), bw.GetPosition())

	// Big Endian
	w = bytes.NewBuffer([]byte{})
	bw = NewWriter(w, binary.BigEndian)

	err = bw.WriteFloat64(float64(1234.5678))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x40, 0x93, 0x4A, 0x45, 0x6D, 0x5C, 0xFA, 0xAD}, w.Bytes())
	assert.Equal(t, int64(8), bw.GetPosition())
}

func TestWriteFloat64Error(t *testing.T) {
	t.Parallel()
	// nil writer
	bw := Writer{}
	bw.bo = binary.LittleEndian
	assert.Error(t, bw.WriteFloat64(float64(1234.5678)))

	// nil byte order
	bw = Writer{}
	bw.dest = bytes.NewBuffer([]byte{})
	assert.Error(t, bw.WriteFloat64(float64(1234.5678)))

	// writer error
	bw = NewWriter(errRW, binary.LittleEndian)
	assert.Error(t, bw.WriteFloat64(float64(1234.5678)))
}

func TestZeroFill(t *testing.T) {
	t.Parallel()
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.LittleEndian)
	assert.NoError(t, bw.ZeroFill(64))
	assert.Equal(t, make([]byte, 64), w.Bytes())
	assert.Equal(t, int64(64), bw.GetPosition())

	// ZeroFill zero length should shortcut
	assert.NoError(t, bw.ZeroFill(0))

	// zero-fill (relatively) large amount of bytes
	expectedPos := int64(64)
	for _, v := range []int64{800, 1024, 1080, 4096, 8000} {
		bw.ZeroFill(v)
		expectedPos += v
		assert.Equal(t, expectedPos, bw.GetPosition())
	}
}

func TestZeroFillError(t *testing.T) {
	t.Parallel()
	// nil writer
	bw := Writer{}
	bw.bo = binary.LittleEndian
	assert.Error(t, bw.ZeroFill(64))

	// writer error
	bw = NewWriter(errRW, binary.LittleEndian)
	assert.Error(t, bw.ZeroFill(64))

	// writer error with large enough discard to cause chunking
	err = bw.ZeroFill(4096)
	assert.Error(t, err)

	// ZeroFill with negative length
	assert.Error(t, bw.ZeroFill(-1))
}

func TestWriterReset(t *testing.T) {
	t.Parallel()
	// Big Endian writer at position 10
	w := bytes.NewBuffer([]byte{})
	bw := NewWriter(w, binary.BigEndian)
	assert.NoError(t, bw.ZeroFill(10))
	assert.Equal(t, int64(10), bw.GetPosition())

	// Reset to Little Endian
	w2 := bytes.NewBuffer([]byte{})
	bw.Reset(w2, binary.LittleEndian)
	assert.Equal(t, int64(0), bw.GetPosition())
	assert.Equal(t, binary.LittleEndian, bw.GetByteOrder())
	assert.NoError(t, bw.ZeroFill(4))
	assert.Equal(t, int64(4), bw.GetPosition())
}

/*
===============================================================================
    baseBinary
===============================================================================
*/

func TestGetPosition(t *testing.T) {
	t.Parallel()
	buf := []byte("1234567890abcdef")
	bb := NewReaderBytes(buf, binary.LittleEndian)
	ui16 := uint16(0)
	ui32 := uint32(0)
	tmp3 := make([]byte, 3)
	c := byte(0)
	assert.Equal(t, int64(0), bb.GetPosition())
	bb.Discard(4)
	assert.Equal(t, int64(4), bb.GetPosition())
	bb.ReadByte(&c)
	assert.Equal(t, int64(5), bb.GetPosition())
	bb.ReadUint16(&ui16)
	assert.Equal(t, int64(7), bb.GetPosition())
	bb.ReadUint32(&ui32)
	assert.Equal(t, int64(11), bb.GetPosition())
	bb.ReadBytes(tmp3)
	assert.Equal(t, int64(14), bb.GetPosition())
}

func TestGetByteOrder(t *testing.T) {
	t.Parallel()
	bb := NewReaderBytes([]byte{}, binary.LittleEndian)
	assert.Equal(t, binary.LittleEndian, bb.GetByteOrder())

	bb = NewReaderBytes([]byte{}, binary.BigEndian)
	assert.Equal(t, binary.BigEndian, bb.GetByteOrder())
}

func TestSetByteOrder(t *testing.T) {
	t.Parallel()
	bb := NewReaderBytes([]byte{}, binary.LittleEndian)
	bb.SetByteOrder(binary.BigEndian)
	assert.Equal(t, binary.BigEndian, bb.GetByteOrder())
	bb.SetByteOrder(binary.LittleEndian)
	assert.Equal(t, binary.LittleEndian, bb.GetByteOrder())
}

// Benchmarks

type devNull int

// devNull implements `io.Reader` and `io.Writer` to remove reader-specific impact on benchmarks
var blackHole = devNull(0)
var buf []byte
var err error
var c byte
var brLE = NewReader(blackHole, binary.LittleEndian)
var brBE = NewReader(blackHole, binary.BigEndian)
var bwLE = NewWriter(blackHole, binary.LittleEndian)
var bwBE = NewWriter(blackHole, binary.BigEndian)

func (devNull) Read(p []byte) (int, error) {
	return len(p), nil
}

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

type errorRW int
type negativeRW int

var errRW = errorRW(0)
var negRW = negativeRW(0)

func (errorRW) Write(p []byte) (int, error) {
	return 0, errors.New("error")
}

func (errorRW) Read(p []byte) (int, error) {
	return 0, errors.New("error")
}

func (negativeRW) Write(p []byte) (int, error) {
	return -len(p), nil
}

func (negativeRW) Read(p []byte) (int, error) {
	return -len(p), nil
}
func BenchmarkReadByte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err = brLE.ReadByte(&c)
		if err != nil {
			b.Fatal(err)
		}
		if c != 0 {
			b.Fatal("c != 0")
		}
	}
}

func BenchmarkReadBytes(b *testing.B) {
	benchmarks := [][]byte{
		make([]byte, 4),
		make([]byte, 32),
		make([]byte, 128),
		make([]byte, 1024),
		make([]byte, 4096),
	}
	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("BenchmarkReadBytes(%d)", len(bm)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err = brLE.ReadBytes(bm)
				if err != nil {
					panic(err)
				}
				if bm[0] != 0 {
					panic("bm[0] != 0")
				}
			}
		})
	}
}

func BenchmarkPeek(b *testing.B) {
	benchmarks := [][]byte{
		make([]byte, 4),
		make([]byte, 128),
		make([]byte, 512),
	}
	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("BenchmarkPeek(%d)", len(bm)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err = brLE.Peek(bm)
				if err != nil {
					panic(err)
				}
				brLE.Reset(blackHole, binary.LittleEndian)
				// to prevent re-use across benchmark runs:
				//brLE.peekBuffer = make([]byte, 64)
				// (but this could be argued to be unrealistic, as the reader will
				// not likely be reset every call to Peek)
			}
		})
	}
}

func BenchmarkPeekThenRead(b *testing.B) {
	pb := make([]byte, 4)
	for i := 0; i < b.N; i++ {
		err := brLE.Peek(pb)
		if err != nil {
			panic(err)
		}
		err = brLE.ReadBytes(pb)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkWriteBytes(b *testing.B) {
	benchmarks := [][]byte{
		make([]byte, 4),
		make([]byte, 32),
		make([]byte, 128),
		make([]byte, 1024),
		make([]byte, 4096),
	}
	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("BenchmarkWriteBytes(%d)", len(bm)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err = bwLE.WriteBytes(bm)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func BenchmarkReadUint16(b *testing.B) {
	ui16 := uint16(9000)
	for i := 0; i < b.N; i++ {
		err = brLE.ReadUint16(&ui16)
		if err != nil {
			panic(err)
		}
		if ui16 != 0 {
			panic("ui16 != 0")
		}
	}
}

func BenchmarkWriteUint16(b *testing.B) {
	ui16 := uint16(9000)
	for i := 0; i < b.N; i++ {
		err = bwLE.WriteUint16(ui16)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkReadUint32(b *testing.B) {
	ui32 := uint32(9000)
	for i := 0; i < b.N; i++ {
		err = brLE.ReadUint32(&ui32)
		if err != nil {
			panic(err)
		}
		if ui32 != 0 {
			panic("ui32 != 0")
		}
	}
}

func BenchmarkReadUint64(b *testing.B) {
	ui64 := uint64(9000)
	for i := 0; i < b.N; i++ {
		err = brLE.ReadUint64(&ui64)
		if err != nil {
			panic(err)
		}
		if ui64 != 0 {
			panic("ui64 != 0")
		}
	}
}

func BenchmarkDiscard(b *testing.B) {
	benchmarks := []int64{
		32,
		1024,
		4096,
		8000,
		18051,
	}
	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("BenchmarkDiscard(%d)", bm), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err = brLE.Discard(bm)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func BenchmarkZeroFill(b *testing.B) {
	benchmarks := []int64{
		32,
		1024,
		4096,
		8000,
		18051,
	}
	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("BenchmarkZeroFill(%d)", bm), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err = bwLE.ZeroFill(bm)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func TestXYZ(t *testing.T) {
	dest := bytes.NewBuffer([]byte{})
	w := NewWriter(dest, binary.LittleEndian)
	w.WriteUint32(uint32(0x32323232))
	w.WriteUint64(uint64(0x6464646464646464))
	out := dest.Bytes()
	assert.Equal(t, uint32(0x32323232), binary.LittleEndian.Uint32(out[0:4]))
	assert.Equal(t, uint64(0x6464646464646464), binary.LittleEndian.Uint64(out[4:12]))
}

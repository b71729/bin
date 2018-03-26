package bin

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testBuffer = []byte("1234567890abcdefghijklmnopqrstuvwxyz")

func TestNewBuffer(t *testing.T) {
	t.Parallel()
	bb := NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	assert.Equal(t, int64(0), bb.GetPosition())
}

func TestReadByte(t *testing.T) {
	t.Parallel()
	bb := NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	c := byte(0)
	for _, expected := range testBuffer {
		err := bb.ReadByte(&c)
		assert.NoError(t, err)
		assert.Equal(t, expected, c)
	}
}
func TestReadByteError(t *testing.T) {
	t.Parallel()
	bb := NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF
	bb.source.Read(make([]byte, len(testBuffer)))
	c := byte(0)
	err := bb.ReadByte(&c)
	assert.Error(t, err)
}

func TestRead(t *testing.T) {
	t.Parallel()
	bb := NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	buf := make([]byte, 32)
	nread, err := bb.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), nread)
	assert.Equal(t, testBuffer[:32], buf[:32])
}

func TestReadBytes(t *testing.T) {
	t.Parallel()
	buf = append(testBuffer, make([]byte, 1024)...)
	bb := NewBinaryReaderBytes(buf, binary.LittleEndian)

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
	bb := BinaryReader{}
	err := bb.ReadBytes(buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadBytes(buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-10))
	err = bb.ReadBytes(buf)
	assert.Error(t, err)
}

func TestReadUint16(t *testing.T) {
	t.Parallel()
	// Little Endian
	buf := []byte{0x08, 0x00, 0xFF, 0x01}
	bb := NewBinaryReaderBytes(buf, binary.LittleEndian)
	ui16 := uint16(0)

	err := bb.ReadUint16(&ui16)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0x0008), ui16)

	err = bb.ReadUint16(&ui16)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0x01FF), ui16)

	// Big Endian
	buf = []byte{0x08, 0x00, 0xFF, 0x01}
	bb = NewBinaryReaderBytes(buf, binary.BigEndian)

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
	bb := BinaryReader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadUint16(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = BinaryReader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadUint16(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadUint16(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-1))
	err = bb.ReadUint16(&buf)
	assert.Error(t, err)
}

func TestReadUint32(t *testing.T) {
	t.Parallel()
	// Little Endian
	buf := []byte{0x08, 0x00, 0xFF, 0x01}
	bb := NewBinaryReaderBytes(buf, binary.LittleEndian)
	ui32 := uint32(0)

	err := bb.ReadUint32(&ui32)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0x01FF0008), ui32)

	// Big Endian
	buf = []byte{0x08, 0x00, 0xFF, 0x01}
	bb = NewBinaryReaderBytes(buf, binary.BigEndian)
	err = bb.ReadUint32(&ui32)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0x0800FF01), ui32)
}

func TestReadUint32Error(t *testing.T) {
	t.Parallel()
	buf := uint32(0)

	// nil reader
	bb := BinaryReader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadUint32(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = BinaryReader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadUint32(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadUint32(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-3))
	err = bb.ReadUint32(&buf)
	assert.Error(t, err)
}

func TestReadUint64(t *testing.T) {
	t.Parallel()
	// Little Endian
	buf := []byte{0x08, 0x00, 0xFF, 0x01, 0x08, 0x00, 0xFF, 0x01}
	bb := NewBinaryReaderBytes(buf, binary.LittleEndian)
	ui64 := uint64(0)

	err := bb.ReadUint64(&ui64)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x01FF000801FF0008), ui64)

	// Big Endian
	buf = []byte{0x08, 0x00, 0xFF, 0x01, 0x08, 0x00, 0xFF, 0x01}
	bb = NewBinaryReaderBytes(buf, binary.BigEndian)
	err = bb.ReadUint64(&ui64)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x800FF010800FF01), ui64)
}

func TestReadUint64Error(t *testing.T) {
	t.Parallel()
	buf := uint64(0)

	// nil reader
	bb := BinaryReader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadUint64(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = BinaryReader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadUint64(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadUint64(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
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
	bb := NewBinaryReaderBytes(buf, binary.LittleEndian)
	err := bb.ReadFloat32(&f32)
	assert.NoError(t, err)
	assert.Equal(t, float32(123.456), f32)

	// Big Endian
	buf = []byte{0x42, 0xf6, 0xe9, 0x79}
	bb = NewBinaryReaderBytes(buf, binary.BigEndian)
	err = bb.ReadFloat32(&f32)
	assert.NoError(t, err)
	assert.Equal(t, float32(123.456), f32)
}

func TestReadFloat32Error(t *testing.T) {
	t.Parallel()
	buf := float32(0)

	// nil reader
	bb := BinaryReader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadFloat32(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = BinaryReader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadFloat32(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadFloat32(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
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
	bb := NewBinaryReaderBytes(buf, binary.LittleEndian)
	err := bb.ReadFloat64(&f64)
	assert.NoError(t, err)
	assert.Equal(t, float64(123.456), f64)

	// Big Endian
	buf = []byte{0x40, 0x5E, 0xDD, 0x2F, 0x1A, 0x9F, 0xBE, 0x77}
	bb = NewBinaryReaderBytes(buf, binary.BigEndian)
	err = bb.ReadFloat64(&f64)
	assert.NoError(t, err)
	assert.Equal(t, float64(123.456), f64)
}

func TestReadFloat64Error(t *testing.T) {
	t.Parallel()
	buf := float64(0)

	// nil reader
	bb := BinaryReader{}
	bb.bo = binary.LittleEndian
	err := bb.ReadFloat64(&buf)
	assert.Error(t, err)

	// nil byte order
	bb = BinaryReader{}
	bb.source = bytes.NewReader(testBuffer)
	err = bb.ReadFloat64(&buf)
	assert.Error(t, err)

	// Reached EOF
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	bb.source.Read(make([]byte, len(testBuffer)))
	err = bb.ReadFloat64(&buf)
	assert.Error(t, err)

	// Reached EOF during read (partial)
	bb = NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	// Reached EOF already
	bb.source.Read(make([]byte, len(testBuffer)-1))
	err = bb.ReadFloat64(&buf)
	assert.Error(t, err)
}

func TestDiscard(t *testing.T) {
	t.Parallel()
	bb := NewBinaryReaderBytes(testBuffer, binary.LittleEndian)
	bb.Discard(0)
	assert.Equal(t, int64(0), bb.GetPosition())
	bb.Discard(1)
	assert.Equal(t, int64(1), bb.GetPosition())
	bb.Discard(7)
	assert.Equal(t, int64(8), bb.GetPosition())
	bb.Discard(2)
	assert.Equal(t, int64(10), bb.GetPosition())
}

func TestDiscardError(t *testing.T) {
	t.Parallel()
	// nil reader
	bb := BinaryReader{}
	bb.bo = binary.LittleEndian
	err := bb.Discard(2)
	assert.Error(t, err)
}

func TestGetPosition(t *testing.T) {
	t.Parallel()
	buf := []byte("1234567890abcdef")
	bb := NewBinaryReaderBytes(buf, binary.LittleEndian)
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
	bb := NewBinaryReaderBytes([]byte{}, binary.LittleEndian)
	assert.Equal(t, binary.LittleEndian, bb.GetByteOrder())

	bb = NewBinaryReaderBytes([]byte{}, binary.BigEndian)
	assert.Equal(t, binary.BigEndian, bb.GetByteOrder())
}

func TestSetByteOrder(t *testing.T) {
	t.Parallel()
	bb := NewBinaryReaderBytes([]byte{}, binary.LittleEndian)
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
var brLE = NewBinaryReader(blackHole, binary.LittleEndian)
var brBE = NewBinaryReader(blackHole, binary.BigEndian)

func (devNull) Read(p []byte) (int, error) {
	return len(p), nil
}

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
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

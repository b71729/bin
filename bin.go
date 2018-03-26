package bin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

/*
===============================================================================
    Data Types
===============================================================================
*/

// BinaryReader provides methods for reading various data types from an `io.Reader`.
type BinaryReader struct {
	binaryBase
	source io.Reader
}

// binaryBase contains a set of methods and variables common to sibling
// interfaces.
type binaryBase struct {
	pos int64
	bo  binary.ByteOrder
	tmpBuffers
}

// tmpBuffers provides an assortment of temporary variables used internally
// to reduce allocation overhead.
//
// These variables are **not** safe for concurrent use; can consider the use
// of Mutex if the need arises.
type tmpBuffers struct {
	b8  [8]byte
	i   int
	err error
}

/*
===============================================================================
    BinaryReader
===============================================================================
*/

// ReadByte reads one byte into `dst`.
func (b *BinaryReader) ReadByte(dst *byte) error {
	b.err = b.ReadBytes(b.b8[:1])
	if b.err != nil {
		return b.err
	}
	*dst = b.b8[0]
	return nil
}

// Read satisfies the Liskov Subsitution Principle of its base `io.Reader`
func (b *BinaryReader) Read(p []byte) (n int, err error) {
	n, err = b.source.Read(p)
	return
}

// ReadBytes attempts to read `len(dst)` bytes into `dst`.
// Multiple calls will be made to `source.Read` in the case of a partial read.
//
// If unable to completely read into `dst`, `io.ErrUnexpectedEOF` will be returned.
func (b *BinaryReader) ReadBytes(dst []byte) error {
	if b.source == nil {
		return errors.New("ReadBytes([]byte): reader is nil")
	}
	// here `io.ReadFull` is used to ensure all requested bytes are read
	// via repeated `Read` calls
	b.i, b.err = io.ReadFull(b.source, dst)
	b.pos += int64(b.i)
	if b.err != nil {
		return b.err
	}
	if b.i < len(dst) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// ReadUint16 reads an unsigned 16-bit integer into `dst` according to the current byte order.
func (b *BinaryReader) ReadUint16(dst *uint16) error {
	if b.source == nil {
		return errors.New("ReadUint16(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadUint16(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b.b8[:2]); b.err != nil {
		return b.err
	}
	*dst = b.bo.Uint16(b.b8[:2])
	return nil
}

// ReadUint32 reads an unsigned 32-bit integer into `dst` according to the current byte order.
func (b *BinaryReader) ReadUint32(dst *uint32) error {
	if b.source == nil {
		return errors.New("ReadUint32(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadUint32(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b.b8[:4]); b.err != nil {
		return b.err
	}
	*dst = b.bo.Uint32(b.b8[:4])
	return nil
}

// ReadUint64 reads an unsigned 64-bit integer into `dst` according to the current byte order.
func (b *BinaryReader) ReadUint64(dst *uint64) error {
	if b.source == nil {
		return errors.New("ReadUint64(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadUint64(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b.b8[:8]); b.err != nil {
		return b.err
	}
	*dst = b.bo.Uint64(b.b8[:8])
	return nil
}

// ReadFloat32 reads a 32-bit IEEE 754 floating-point integer into `dst`
// according to the current byte order.
func (b *BinaryReader) ReadFloat32(dst *float32) error {
	if b.source == nil {
		return errors.New("ReadFloat32(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadFloat32(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b.b8[:4]); b.err != nil {
		return b.err
	}
	*dst = math.Float32frombits(b.bo.Uint32(b.b8[:4]))
	return nil
}

// ReadFloat64 reads a 64-bit IEEE 754 floating-point integer into `dst`
// according to the current byte order.
func (b *BinaryReader) ReadFloat64(dst *float64) error {
	if b.source == nil {
		return errors.New("ReadFloat64(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadFloat64(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b.b8[:8]); b.err != nil {
		return b.err
	}
	*dst = math.Float64frombits(b.bo.Uint64(b.b8[:8]))
	return nil
}

// Discard reads `n` bytes into a discarded buffer. This could use optimisation.
func (b *BinaryReader) Discard(n int64) error {
	// NOTE: Expensive. Needs improving.
	if b.source == nil {
		return fmt.Errorf("Discard(%d): reader is nil", n)
	}
	return b.ReadBytes(make([]byte, n))
}

// NewBinaryReader creates a new `BinaryReader` encapsulating the given `source`,
// and using the byte order `bo` to specify endianness.
//
// For futureproofing, it is suggested to use these constructors rather than
// manually creating an instance (i.e. `br := BinaryReader{}`)
func NewBinaryReader(source io.Reader, bo binary.ByteOrder) (br BinaryReader) {
	br.source = source
	br.bo = bo
	return
}

// NewBinaryReaderBytes creates a new `BinaryReader` to read from the given `source`,
// using the byte order `bo` to specify endianness.
//
// Since `source` is a slice of bytes, a `bytes.Reader` will wrap the slice to satisfy
// the `io.Reader` interface.
//
// For futureproofing, it is suggested to use these constructors rather than
// manually creating an instance (i.e. `br := BinaryReader{}`)
func NewBinaryReaderBytes(source []byte, bo binary.ByteOrder) (br BinaryReader) {
	br.source = bytes.NewReader(source)
	br.bo = bo
	return
}

// GetPosition returns the current reader offset as a 64-bit integer.
func (b *binaryBase) GetPosition() int64 {
	return b.pos
}

// SetByteOrder sets the current byte order to `bo`.
// This can be done on-the-fly.
func (b *binaryBase) SetByteOrder(bo binary.ByteOrder) {
	b.bo = bo
}

// GetByteOrder returns the current byte order.
//
// Note that this can be `nil` if the interface was not created via a
// constructor method.
func (b *binaryBase) GetByteOrder() binary.ByteOrder {
	return b.bo
}

// Package bin provides utility interfaces for working with binary data streams
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

// Reader provides methods for reading various data types from an `io.Reader`.
type Reader struct {
	binaryBase
	source     io.Reader
	peekBuffer []byte // this is initially set to 64 bytes
	nPeeked    int
	peekPos    int
}

// Writer provides methods for reading various data types to an `io.Writer`.
type Writer struct {
	binaryBase
	dest io.Writer
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
	_1kb    [1024]byte
	null1kb [1024]byte // do not write to this
	i       int
	i64     int64
	err     error
}

/*
===============================================================================
    Reader
===============================================================================
*/

// ReadByte reads one byte into `dst`.
func (b *Reader) ReadByte(dst *byte) error {
	if b.err = b.ReadBytes(b._1kb[:1]); b.err != nil {
		return b.err
	}
	*dst = b._1kb[0]
	return nil
}

// Read satisfies the Liskov Subsitution Principle of its base `io.Reader`
func (b *Reader) Read(p []byte) (n int, err error) {
	n, err = b.source.Read(p)
	b.pos += int64(n)
	return
}

// ReadBytes attempts to read `len(dst)` bytes into `dst`.
// Multiple calls will be made to `source.Read` in the case of a partial read.
//
// If unable to completely read into `dst`, `io.ErrUnexpectedEOF` will be returned.
func (b *Reader) ReadBytes(dst []byte) error {
	if b.source == nil {
		return fmt.Errorf("ReadBytes([%d]byte): reader is nil", len(dst))
	}
	// shortcut if `dst` has a length of zero
	if len(dst) == 0 {
		return nil
	}
	if b.numUnusedPeekedBytes() > 0 {
		b.i = len(dst)
		// we have peeked some bytes.
		// therefore, they should be used before remaining bytes in reader.
		if b.numUnusedPeekedBytes() >= b.i {
			// unused bytes in peek buffer exceeed the destination length,
			// so use up what we can
			copy(dst, b.peekBuffer[b.peekPos:b.peekPos+b.i])
			b.peekPos += b.i
			b.err = nil
		} else {
			// more bytes are requested than available in peek buffer
			// fulfill partially from peek buffer, then `io.ReadFull` the rest
			copy(dst, b.peekBuffer[b.peekPos:b.nPeeked])

			b.i, b.err = io.ReadFull(b.source, dst[(b.nPeeked-b.peekPos):])

			// also advance reader position by those bytes we used
			b.i += (b.nPeeked - b.peekPos)
			// b.nPeeked = 0
			// b.peekPos = 0
			// TODO: Is above better?
			b.peekPos = b.nPeeked

		}
	} else {
		// here `io.ReadFull` is used to ensure all requested bytes are read
		// via repeated `Read` calls
		b.i, b.err = io.ReadFull(b.source, dst)
	}

	b.pos += int64(b.i)
	return b.err
}

// ReadUint16 reads an unsigned 16-bit integer into `dst` according to the current byte order.
func (b *Reader) ReadUint16(dst *uint16) error {
	if b.source == nil {
		return errors.New("ReadUint16(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadUint16(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b._1kb[:2]); b.err != nil {
		return b.err
	}
	*dst = b.bo.Uint16(b._1kb[:2])
	return nil
}

// ReadUint32 reads an unsigned 32-bit integer into `dst` according to the current byte order.
func (b *Reader) ReadUint32(dst *uint32) error {
	if b.source == nil {
		return errors.New("ReadUint32(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadUint32(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b._1kb[:4]); b.err != nil {
		return b.err
	}
	*dst = b.bo.Uint32(b._1kb[:4])
	return nil
}

// ReadUint64 reads an unsigned 64-bit integer into `dst` according to the current byte order.
func (b *Reader) ReadUint64(dst *uint64) error {
	if b.source == nil {
		return errors.New("ReadUint64(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadUint64(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b._1kb[:8]); b.err != nil {
		return b.err
	}
	*dst = b.bo.Uint64(b._1kb[:8])
	return nil
}

// ReadFloat32 reads a 32-bit IEEE 754 floating-point integer into `dst`
// according to the current byte order.
func (b *Reader) ReadFloat32(dst *float32) error {
	if b.source == nil {
		return errors.New("ReadFloat32(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadFloat32(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b._1kb[:4]); b.err != nil {
		return b.err
	}
	*dst = math.Float32frombits(b.bo.Uint32(b._1kb[:4]))
	return nil
}

// ReadFloat64 reads a 64-bit IEEE 754 floating-point integer into `dst`
// according to the current byte order.
func (b *Reader) ReadFloat64(dst *float64) error {
	if b.source == nil {
		return errors.New("ReadFloat64(): reader is nil")
	}
	if b.bo == nil {
		return errors.New("ReadFloat64(): ByteOrder is not set")
	}
	if b.err = b.ReadBytes(b._1kb[:8]); b.err != nil {
		return b.err
	}
	*dst = math.Float64frombits(b.bo.Uint64(b._1kb[:8]))
	return nil
}

// Discard reads `n` bytes into a discarded buffer.
func (b *Reader) Discard(n int64) error {
	if b.source == nil {
		return fmt.Errorf("Discard(%d): reader is nil", n)
	}
	b.i64 = n
	if b.i64 <= 1024 { // shortcut
		return b.ReadBytes(b._1kb[:n])
	}
	// cut away at `n` until we have <= 1024 bytes remaining to discard
	for b.i64 > 1024 {
		if b.err = b.ReadBytes(b._1kb[:]); b.err != nil {
			return b.err
		}
		b.i64 -= 1024
	}
	// and then discard the rest
	// this function should have caused zero allocs.
	return b.ReadBytes(b._1kb[:b.i64])
}

// numUnusedPeekedBytes returns the number of bytes that have been peeked
// but not consumed in a subsequent call to the reader.
func (b *Reader) numUnusedPeekedBytes() int {
	return b.nPeeked - b.peekPos
}

// Peek returns the next n bytes without advancing the reader.
// If the operation cannot fully write to `dst`, it will return an error.
func (b *Reader) Peek(dst []byte) error {
	if b.source == nil {
		return fmt.Errorf("Peek([%d]byte): reader is nil", len(dst))
	}
	b.i = len(dst)
	// shortcut if `dst` has a length of zero
	if b.i == 0 {
		return nil
	}
	// we may have already peeked bytes
	if b.numUnusedPeekedBytes() > 0 {

		// we have peeked some bytes.
		// therefore, they should be used before reading from `source`
		if b.numUnusedPeekedBytes() >= b.i {
			// unused bytes in peek buffer exceeed the destination length,
			// so use up what we can
			copy(dst, b.peekBuffer[b.peekPos:b.peekPos+b.i])
			return nil
		}
		// more bytes are requested than available in peek buffer
		// fulfill partially from peek buffer, then `io.ReadFull` the rest
		copy(dst, b.peekBuffer[b.peekPos:b.nPeeked])
	}
	nRead := b.i - b.numUnusedPeekedBytes()

	// determine whether we need to grow buffer
	b.i = ((b.nPeeked + nRead) - len(b.peekBuffer)) // if > 0 then we need to grow

	if b.i > 0 {
		// grow by a minimum of 1kb to reduce overhead
		if b.i <= 64 {
			b.peekBuffer = append(b.peekBuffer, make([]byte, 64)...)
		} else {
			b.peekBuffer = append(b.peekBuffer, make([]byte, b.i)...)
		}
	}

	if _, b.err = io.ReadFull(b.source, b.peekBuffer[b.nPeeked:b.nPeeked+nRead]); b.err != nil {
		return b.err
	}

	copy(dst[b.numUnusedPeekedBytes():], b.peekBuffer[b.nPeeked:b.nPeeked+nRead])
	b.nPeeked += nRead
	return nil
}

// Reset resets the reader position and source `io.Reader` to `source`
func (b *Reader) Reset(source io.Reader, bo binary.ByteOrder) {
	b.pos = 0
	b.source = source
	b.bo = bo
	b.peekPos = 0
	b.nPeeked = 0
}

// NewReader creates a new `Reader` encapsulating the given `source`,
// and using the byte order `bo` to specify endianness.
//
// For futureproofing, it is suggested to use these constructors rather than
// manually creating an instance (i.e. `br := Reader{}`)
func NewReader(source io.Reader, bo binary.ByteOrder) Reader {
	br := Reader{
		source:     source,
		peekBuffer: make([]byte, 64),
	}
	br.bo = bo
	return br
}

// NewReaderBytes creates a new `Reader` to read from the given `source`,
// using the byte order `bo` to specify endianness.
//
// Since `source` is a slice of bytes, a `bytes.Reader` will wrap the slice to satisfy
// the `io.Reader` interface.
//
// For futureproofing, it is suggested to use these constructors rather than
// manually creating an instance (i.e. `br := Reader{}`)
func NewReaderBytes(source []byte, bo binary.ByteOrder) Reader {
	return NewReader(bytes.NewReader(source), bo)
}

/*
===============================================================================
    Writer
===============================================================================
*/

// WriteByte writes a byte
func (b *Writer) WriteByte(src byte) error {
	if b.dest == nil {
		return fmt.Errorf("WriteByte(%02X): writer is nil", src)
	}
	b._1kb[0] = src
	return b.WriteBytes(b._1kb[:1])
}

// Write satisfies the Liskov Subsitution Principle of its base `io.Writer`
func (b *Writer) Write(p []byte) (n int, err error) {
	// shortcut if `p` has a length of zero
	if len(p) == 0 {
		return 0, nil
	}
	n, err = b.dest.Write(p)
	b.pos += int64(n)
	return
}

// WriteBytes writes all bytes from `src`
func (b *Writer) WriteBytes(src []byte) error {
	if b.dest == nil {
		return errors.New("WriteBytes([]byte): writer is nil")
	}
	// shortcut if `src` has a length of zero
	if len(src) == 0 {
		return nil
	}
	b.i, b.err = b.dest.Write(src)
	b.pos += int64(b.i)
	if b.err != nil {
		return b.err
	}
	return nil
}

// WriteUint16 writes an unsigned 16-bit integer according to the current byte order.
func (b *Writer) WriteUint16(src uint16) error {
	if b.dest == nil {
		return fmt.Errorf("WriteUint16(%08X): writer is nil", src)
	}
	if b.bo == nil {
		return fmt.Errorf("WriteUint16(%08X): ByteOrder is not set", src)
	}
	b.bo.PutUint16(b._1kb[:2], src)
	return b.WriteBytes(b._1kb[:2])
}

// WriteUint32 writes an unsigned 32-bit integer according to the current byte order.
func (b *Writer) WriteUint32(src uint32) error {
	if b.dest == nil {
		return fmt.Errorf("WriteUint32(%016X): writer is nil", src)
	}
	if b.bo == nil {
		return fmt.Errorf("WriteUint32(%016X): ByteOrder is not set", src)
	}
	b.bo.PutUint32(b._1kb[:4], src)
	return b.WriteBytes(b._1kb[:4])
}

// WriteUint64 writes an unsigned 64-bit integer according to the current byte order.
func (b *Writer) WriteUint64(src uint64) error {
	if b.dest == nil {
		return fmt.Errorf("WriteUint64(%032X): writer is nil", src)
	}
	if b.bo == nil {
		return fmt.Errorf("WriteUint64(%032X): ByteOrder is not set", src)
	}
	b.bo.PutUint64(b._1kb[:8], src)
	return b.WriteBytes(b._1kb[:8])
}

// WriteFloat32 writes a 32-bit IEEE 754 floating-point integer
// according to the current byte order.
func (b *Writer) WriteFloat32(src float32) error {
	if b.dest == nil {
		return fmt.Errorf("WriteFloat32(%f): writer is nil", src)
	}
	if b.bo == nil {
		return fmt.Errorf("WriteFloat32(%f): ByteOrder is not set", src)
	}
	b.bo.PutUint32(b._1kb[:4], math.Float32bits(src))
	return b.WriteBytes(b._1kb[:4])
}

// WriteFloat64 writes a 64-bit IEEE 754 floating-point integer
// according to the current byte order.
func (b *Writer) WriteFloat64(src float64) error {
	if b.dest == nil {
		return fmt.Errorf("WriteFloat64(%f): writer is nil", src)
	}
	if b.bo == nil {
		return fmt.Errorf("WriteFloat64(%f): ByteOrder is not set", src)
	}
	b.bo.PutUint64(b._1kb[:8], math.Float64bits(src))
	return b.WriteBytes(b._1kb[:8])
}

// ZeroFill writes `n` null-bytes.
func (b *Writer) ZeroFill(n int64) error {
	if b.dest == nil {
		return fmt.Errorf("ZeroFill(%d): writer is nil", n)
	}
	// cleanse input to procedure
	if n < 0 {
		return fmt.Errorf("ZeroFill(%d): negative length", n)
	}
	// shortcut if `n` is zero
	if n == 0 {
		return nil
	}
	b.i64 = n
	if b.i64 <= 1024 { // shortcut
		return b.WriteBytes(b.null1kb[:n])
	}
	// cut away at `n` until we have <= 1024 bytes remaining to fill
	for b.i64 > 1024 {
		if b.err = b.WriteBytes(b.null1kb[:]); b.err != nil {
			return b.err
		}
		b.i64 -= 1024
	}
	// and then discard the rest
	// this function should have caused zero allocs.
	return b.WriteBytes(b.null1kb[:b.i64])

}

// Reset resets the writer position and source `io.Writer` to `dest`
func (b *Writer) Reset(dest io.Writer, bo binary.ByteOrder) {
	b.pos = 0
	b.dest = dest
	b.bo = bo
}

// NewWriter creates a new `Writer` targetted at the given `dest`,
// and using the byte order `bo` to specify endianness.
//
// For futureproofing, it is suggested to use these constructors rather than
// manually creating an instance (i.e. `bw := Writer{}`)
func NewWriter(dest io.Writer, bo binary.ByteOrder) (bw Writer) {
	bw.dest = dest
	bw.bo = bo
	return
}

/*
===============================================================================
    binaryBase
===============================================================================
*/

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

package pdb

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// TypeRecord records information about a type.
type TypeRecord struct {
	Hdr *TypeRecordHeader
	// TODO: add fields. potentially turn into an interface?
}

// parseTypeRecord parses the given type record, reading from r.
func (file *File) parseTypeRecord(r io.Reader) (TypeRecord, error) {
	// TypeRecordHeader.
	typ := TypeRecord{}
	hdr, err := file.parseTypeRecordHeader(r)
	if err != nil {
		return TypeRecord{}, errors.WithStack(err)
	}
	typ.Hdr = hdr
	// Read type record body contents.
	bodySize := typ.Hdr.RecordSize - 2
	hdr.body = make([]byte, bodySize)
	if _, err := io.ReadFull(r, hdr.body); err != nil {
		return TypeRecord{}, nil
	}
	// TODO: parse type record body.
	return typ, nil
}

// TypeRecordHeader is a type record header.
type TypeRecordHeader struct {
	// Size in bytes of type record, excluding the 2 bytes that make up the size
	// field.
	RecordSize uint16
	RecordKind TypeRecordKind

	body []byte // TODO: remove.
}

// TypeRecordKind denotes the kind of a type record.
//
// ref: LEAF_ENUM_e
type TypeRecordKind uint16

// Type record kinds.
const (
	TypeRecordKindNone TypeRecordKind = 0x0000

	// leaf indices starting records but referenced from symbol records
	TypeRecordKindPointer TypeRecordKind = 0x1002
)

// parseTypeRecordHeader parses the given type record header, reading from r.
func (file *File) parseTypeRecordHeader(r io.Reader) (*TypeRecordHeader, error) {
	// RecordSize.
	hdr := &TypeRecordHeader{}
	if err := binary.Read(r, binary.LittleEndian, &hdr.RecordSize); err != nil {
		return nil, errors.WithStack(err)
	}
	// RecordKind.
	if err := binary.Read(r, binary.LittleEndian, &hdr.RecordKind); err != nil {
		return nil, errors.WithStack(err)
	}
	return hdr, nil
}

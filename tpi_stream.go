package pdb

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/kr/pretty"
	"github.com/pkg/errors"
)

// TPIStream records information about types used in the program. Types are
// referenced by their type index from other parts of the PDB.
//
// ref: https://llvm.org/docs/PDB/TpiStream.html
type TPIStream struct {
	// TPI stream header.
	Hdr *TPIStreamHeader16
	// Type records.
	Types []TypeRecord
}

// parseTPIStream parses the given TPI stream.
func (file *File) parseTPIStream(r io.Reader) (*TPIStream, error) {
	// Parse TPI stream header.
	tpiStream := &TPIStream{}
	hdr, err := file.parseTPIStreamHeader16(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tpiStream.Hdr = hdr
	// Skip padding.
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, &TPIStreamHeader16{})
	hdrSize := int64(buf.Len())
	dbg.Println("   hdrSize:", hdrSize)
	npad := hdrSize % 4
	dbg.Println("   npad:", npad)
	if _, err := io.CopyN(ioutil.Discard, r, npad); err != nil {
		return nil, errors.WithStack(err)
	}
	// Parse type records.
	typeRecordsData := make([]byte, hdr.TypeRecordsSize)
	if _, err := io.ReadFull(r, typeRecordsData); err != nil {
		return nil, errors.WithStack(err)
	}
	dbg.Print("type records data:\n", hex.Dump(typeRecordsData))
	rr := bytes.NewReader(typeRecordsData)
	ntypes := int(hdr.LastTypeID - hdr.FirstTypeID) // TODO: handle 0 case.
	tpiStream.Types = make([]TypeRecord, ntypes)
	for i := 0; i < ntypes; i++ {
		t, err := file.parseTypeRecord(rr)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		pretty.Println("t:", t)
		tpiStream.Types[i] = t
	}
	return tpiStream, nil
}

// TPIStreamHeader16 is a header of the TPI stream with 16-bit type IDs.
//
// ref: HDR_16t in PDB/dbi/tpi.h
type TPIStreamHeader16 struct {
	// TPI version.
	Version TPIVersion
	// First type index, inclusive; type index of first type record in the TPI
	// stream.
	FirstTypeID TypeID16
	// Last type index, exclusive.
	LastTypeID TypeID16
	// Size in bytes of type records data following header.
	TypeRecordsSize int32
	// Index of TPI hash stream.
	HashStreamNum StreamNumber
}

//go:generate stringer -linecomment -type TPIVersion

// TPIVersion specifies the version of Visual Studio Code used to produce the
// TPI. However, in practise, VC80 is almost always used (even if the version of
// VC used to produce the TPI was newer).
type TPIVersion uint32

// TPI versions.
//
// ref: TPIIMPV
const (
	TPIVersionV40        TPIVersion = 19950410 // V 4.0 (1995-04-10)
	TPIVersionV41        TPIVersion = 19951122 // V 4.1 (1995-11-22)
	TPIVersionV50Interim TPIVersion = 19960307 // V 5.0 - interim (1996-03-07)
	TPIVersionV50        TPIVersion = 19961031 // V 5.0 (1996-10-31)
	TPIVersionV70        TPIVersion = 19990903 // V 7.0 (1999-09-03)
	TPIVersionV80        TPIVersion = 20040203 // V 8.0 (2004-02-03)
)

// parseTPIStreamHeader16 parses the given TPI stream header with 16-bit type
// indices.
func (file *File) parseTPIStreamHeader16(r io.Reader) (*TPIStreamHeader16, error) {
	// Version.
	hdr := &TPIStreamHeader16{}
	if err := binary.Read(r, binary.LittleEndian, &hdr.Version); err != nil {
		return nil, errors.WithStack(err)
	}
	// FirstTypeID.
	if err := binary.Read(r, binary.LittleEndian, &hdr.FirstTypeID); err != nil {
		return nil, errors.WithStack(err)
	}
	// LastTypeID.
	if err := binary.Read(r, binary.LittleEndian, &hdr.LastTypeID); err != nil {
		return nil, errors.WithStack(err)
	}
	// TypeRecordsSize.
	if err := binary.Read(r, binary.LittleEndian, &hdr.TypeRecordsSize); err != nil {
		return nil, errors.WithStack(err)
	}
	// HashStreamNum.
	if err := binary.Read(r, binary.LittleEndian, &hdr.HashStreamNum); err != nil {
		return nil, errors.WithStack(err)
	}
	return hdr, nil
}

// TypeID16 is a 16-bit type index which uniquely identifies a type of the PDB.
//
// Any typeID >= Hdr.FirstTypeID is persumed to come from the corresponding TPI
// (or IPI) stream.
//
// A typeID < Hdr.FirstTypeID is decomposed as follows:
//
//    +----------+----------+------------------+
//    | unused   | mode     | kind             |
//    +----------+----------+------------------+
//    |+16       |+12       |+8                |+0
//
// A basic type composed of a type kind and a type mode.
//
//    BasicType = TypeKind | TypeMode
//
//    0b0000MMMMKKKKKKKK // mode and kind bits marked with 'M' and 'K' respectively.
//
// ref [2]: https://llvm.org/docs/PDB/TpiStream.html#type-indices
type TypeID16 uint16

// String returns the string representation of the given basic type.
func (typeID TypeID16) String() string {
	// "The value of the type index for the first type record from the TPI stream
	// is given by the TypeIndexBegin member of the TPI Stream Header although in
	// practice this value is always equal to 0x1000 (4096)" [2].
	if typeID >= 0x1000 {
		return fmt.Sprintf("TypeID(%d)", uint16(typeID))
	}
	mode := TypeMode(typeID & 0x0F00)
	kind := TypeKind(typeID & 0x00FF)
	if mode != TypeModeNone {
		return fmt.Sprintf("%v to %v", mode, kind) // e.g. 64 bit pointer to void
	}
	return kind.String()
}

//go:generate stringer -linecomment -type TypeMode

// TypeMode specifies the mode of basic types (e.g. 32-bit pointer, 64-bit
// pointer).
//
//    0b0000
//    0b0010 - 16 bit pointer
//    0b0011 - 16:16 far pointer
//    0b0100 - 16:16 huge pointer
//    0b0101 - 32 bit pointer
//    0b0110 - 16:32 pointer
//    0b0111 - 64 bit pointer
//
//    0b0000MMMMKKKKKKKK // mode and kind bits marked with 'M' and 'K' respectively.
type TypeMode uint16 // actually 4 bits.

// Type modes.
//
// ref: TYPE_ENUM_e
const (
	TypeModeNone TypeMode = 0x0000 // TypeMode(none)
	// near pointer
	TypeModePointer16 TypeMode = 0x0100 // 16 bit pointer
	// far pointer
	TypeModePointer16Far  TypeMode = 0x0200 // 16:16 far pointer pointer
	TypeModePointer16Huge TypeMode = 0x0300 // 16:16 huge pointer pointer
	TypeModePointer32     TypeMode = 0x0400 // 32 bit pointer
	TypeModePointer32Far  TypeMode = 0x0500 // 16:32 far pointer
	TypeModePointer64     TypeMode = 0x0600 // 64 bit pointer
	TypeModePointer128    TypeMode = 0x0700 // 128 bit pointer
)

//go:generate stringer -linecomment -type TypeKind

// TypeKind specifies the kind of basic types (e.g. int8, uint8).
//
//    0b0000MMMMKKKKKKKK // mode and kind bits marked with 'M' and 'K' respectively.
type TypeKind uint16 // actually 8 bits.

// Type kinds.
//
// ref: TYPE_ENUM_e
const (
	// Special Types
	TypeKindNone            TypeKind = 0x0000 // uncharacterized type (no type)
	TypeKindAbs             TypeKind = 0x0001 // absolute symbol
	TypeKindSegment         TypeKind = 0x0002 // segment type
	TypeKindVoid            TypeKind = 0x0003 // void
	TypeKindHResult         TypeKind = 0x0008 // HRESULT
	TypeKindCurrency        TypeKind = 0x0004 // BASIC 8 byte currency value
	TypeKindBasicStringNear TypeKind = 0x0005 // near BASIC string
	TypeKindBasicStringFar  TypeKind = 0x0006 // far BASIC string
	TypeKindNotTranslated   TypeKind = 0x0007 // type not translated by cvpack
	TypeKindBit             TypeKind = 0x0060 // bit
	TypeKindPascalChar      TypeKind = 0x0061 // Pascal CHAR
	TypeKindBool32FFFFFFFF  TypeKind = 0x0062 // 32-bit BOOL where true is 0xffffffff

	// Character types
	TypeKindCharacter     TypeKind = 0x0070 // really a char
	TypeKindWideCharacter TypeKind = 0x0071 // wide char

	// unicode char types
	TypeKindRune16 TypeKind = 0x007A // 16-bit unicode char
	TypeKindRune32 TypeKind = 0x007B // 32-bit unicode char

	// int types
	TypeKindInt8    TypeKind = 0x0068 // 8 bit signed int
	TypeKindUint8   TypeKind = 0x0069 // 8 bit unsigned int
	TypeKindInt16   TypeKind = 0x0072 // 16 bit signed int
	TypeKindUint16  TypeKind = 0x0073 // 16 bit unsigned int
	TypeKindInt32   TypeKind = 0x0074 // 32 bit signed int
	TypeKindUint32  TypeKind = 0x0075 // 32 bit unsigned int
	TypeKindInt64   TypeKind = 0x0076 // 64 bit signed int
	TypeKindUint64  TypeKind = 0x0077 // 64 bit unsigned int
	TypeKindInt128  TypeKind = 0x0078 // 128 bit signed int
	TypeKindUint128 TypeKind = 0x0079 // 128 bit unsigned int

	// 8 bit character types
	TypeKindInt8Byte  TypeKind = 0x0010 // 8 bit signed
	TypeKindUint8Byte TypeKind = 0x0020 // 8 bit unsigned

	// 16 bit short types
	TypeKindInt16Short  TypeKind = 0x0011 // 16 bit signed
	TypeKindUint16Short TypeKind = 0x0021 // 16 bit unsigned

	// 32 bit long types
	TypeKindInt32Long TypeKind = 0x0012 // 32 bit signed
	// NOTE: by consistency, 0x0022 should be "32 bit unsigned"
	//TypeKindUint32Long TypeKind = 0x0022 // 32 bit unsigned

	// 64 bit quad types
	TypeKindInt64Quad  TypeKind = 0x0013 // 64 bit signed
	TypeKindUint64Quad TypeKind = 0x0023 // 64 bit unsigned

	// 128 bit octet types
	TypeKindInt128Octet  TypeKind = 0x0014 // 128 bit signed
	TypeKindUint128Octet TypeKind = 0x0024 // 128 bit unsigned

	// floating-point types
	TypeKindFloat16   TypeKind = 0x0046 // 16 bit real
	TypeKindFloat32   TypeKind = 0x0040 // 32 bit real
	TypeKindFloat32PP TypeKind = 0x0045 // 32 bit partial-precision real
	TypeKindFloat48   TypeKind = 0x0044 // 48 bit real
	TypeKindFloat64   TypeKind = 0x0041 // 64 bit real
	TypeKindFloat80   TypeKind = 0x0042 // 80 bit real
	TypeKindFloat128  TypeKind = 0x0043 // 128 bit real

	// complex types
	TypeKindComplex32  TypeKind = 0x0050 // 32 bit complex
	TypeKindComplex64  TypeKind = 0x0051 // 64 bit complex
	TypeKindComplex80  TypeKind = 0x0052 // 80 bit complex
	TypeKindComplex128 TypeKind = 0x0053 // 128 bit complex

	// boolean types
	TypeKindBool8   TypeKind = 0x0030 // 8 bit boolean
	TypeKindBool16  TypeKind = 0x0031 // 16 bit boolean
	TypeKindBool32  TypeKind = 0x0032 // 32 bit boolean
	TypeKindBool64  TypeKind = 0x0033 // 64 bit boolean
	TypeKindBool128 TypeKind = 0x0034 // 128 bit boolean

	// ???
	TypeKindInternal TypeKind = 0x00F0 // CV internal type
)

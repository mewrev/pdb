package pdb

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// TPIStream records information about types used in the program. Types are
// referenced by their type index from other parts of the PDB.
//
// ref: https://llvm.org/docs/PDB/TPIStream.html
type TPIStream struct {
	// TPI stream header.
	Hdr *TPIStreamHeader
	// Type records.
	//Types []TypeRecord // TODO: uncomment
}

// parseTPIStream parses the given TPI stream.
func (file *File) parseTPIStream(r io.Reader) (*TPIStream, error) {
	// Parse TPI stream header.
	tpiStream := &TPIStream{}
	hdr, err := file.parseTPIStreamHeader(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tpiStream.Hdr = hdr
	// TODO: parse type records.
	return tpiStream, nil
}

// TPIStreamHeader is a header of the TPI stream.
type TPIStreamHeader struct {
}

// parseTPIStreamHeader parses the given TPI stream header.
func (file *File) parseTPIStreamHeader(r io.Reader) (*TPIStreamHeader, error) {
	hdr := &TPIStreamHeader{}
	// TODO: implement
	return hdr, nil
}

// TypeID is a type index which uniquely identifies a type of the PDB.
//
// Any typeID >= Hdr.FirstTypeID is persumed to come from the corresponding TPI
// (or IPI) stream.
//
// A typeID < Hdr.FirstTypeID is decomposed as follows:
//
//    +----------------------+------+----------+
//    | unused               | mode | kind     |
//    +----------------------+------+----------+
//    |+32                   |+12   |+8        |+0
//
// A basic type composed of a type kind and a type mode.
//
//    BasicType = TypeKind | TypeMode
//
//    0b0000MMMMKKKKKKKK // mode and kind bits marked with 'M' and 'K' respectively.
//
// ref: https://llvm.org/docs/PDB/TpiStream.html#type-indices
type TypeID uint32

// String returns the string representation of the given basic type.
func (typeID TypeID) String() string {
	if typeID >= 0x1000 {
		return fmt.Sprintf("TypeID(%d)", uint32(typeID))
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
	TypeKindHRESULT         TypeKind = 0x0008 // HRESULT
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
	TypeKindChar  TypeKind = 0x0010 // 8 bit signed
	TypeKindUChar TypeKind = 0x0020 // 8 bit unsigned

	// 16 bit short types
	TypeKindShort  TypeKind = 0x0011 // 16 bit signed
	TypeKindUShort TypeKind = 0x0021 // 16 bit unsigned

	// 32 bit long types
	TypeKindLong TypeKind = 0x0012 // 32 bit signed
	// NOTE: by consistency, 0x0022 should be "32 bit unsigned"

	// 64 bit quad types
	TypeKindQuad  TypeKind = 0x0013 // 64 bit signed
	TypeKindUQuad TypeKind = 0x0023 // 64 bit unsigned

	// 128 bit octet types
	TypeKindOctet  TypeKind = 0x0014 // 128 bit signed
	TypeKindUOctet TypeKind = 0x0024 // 128 bit unsigned

	// real types
	TypeKindReal16   TypeKind = 0x0046 // 16 bit real
	TypeKindReal32   TypeKind = 0x0040 // 32 bit real
	TypeKindReal32PP TypeKind = 0x0045 // 32 bit partial-precision real
	TypeKindReal48   TypeKind = 0x0044 // 48 bit real
	TypeKindReal64   TypeKind = 0x0041 // 64 bit real
	TypeKindReal80   TypeKind = 0x0042 // 80 bit real
	TypeKindReal128  TypeKind = 0x0043 // 128 bit real

	// complex types
	TypeKindComplex32  TypeKind = 0x0050 // 32 bit complex
	TypeKindComplex64  TypeKind = 0x0051 // 64 bit complex
	TypeKindComplex80  TypeKind = 0x0052 // 80 bit complex
	TypeKindComplex128 TypeKind = 0x0053 // 128 bit complex

	// boolean types
	TypeKindBool8  TypeKind = 0x0030 // 8 bit boolean
	TypeKindBool16 TypeKind = 0x0031 // 16 bit boolean
	TypeKindBool32 TypeKind = 0x0032 // 32 bit boolean
	TypeKindBool64 TypeKind = 0x0033 // 64 bit boolean

	// ???
	TypeKindInternal TypeKind = 0x00F0 // CV internal type
)

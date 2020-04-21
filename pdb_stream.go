package pdb

import (
	"encoding/binary"
	"io"
	"time"

	"github.com/pkg/errors"
)

// PDBStream records information about the PDB.
//
// ref: https://llvm.org/docs/PDB/PdbStream.html
type PDBStream struct {
	// PDB stream header.
	Hdr *PDBStreamHeader
	// Map from stream name to stream number.
	StreamNameMap *StreamNameMap
}

// parsePDBStream parses the given PDB stream.
func (file *File) parsePDBStream(r io.Reader) (*PDBStream, error) {
	// Parse PDB stream header.
	pdbStream := &PDBStream{}
	hdr, err := file.parsePDBStreamHeader(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pdbStream.Hdr = hdr
	// TODO: parse StreamNameMap.
	return pdbStream, nil
}

// PDBStreamHeader is a header of the PDB stream.
type PDBStreamHeader struct {
	// PDB version.
	Version PDBVersion
	// Creation date.
	Date time.Time
	// Number of times the PDB file as been written to.
	Age uint32
	// Unique ID of the PDB.
	UniqueID GUID
}

//go:generate stringer -linecomment -type PDBVersion

// PDBVersion specifies the version of Visual Studio Code used to produce the
// PDB. However, in practise, VC70 is almost always used (even if the version of
// VC used to produce the PDB was newer).
type PDBVersion uint32

// PDB versions.
//
// ref: PDBIMPV
const (
	PDBVersionVC2     PDBVersion = 19941610 // VC 2 (1994-10-16)
	PDBVersionVC4     PDBVersion = 19950623 // VC 4 (1995-06-23)
	PDBVersionVC41    PDBVersion = 19950814 // VC 4.1 (1995-08-14)
	PDBVersionVC50    PDBVersion = 19960307 // VC 5.0 (1996-03-07)
	PDBVersionVC98    PDBVersion = 19970604 // VC 98 (1997-06-04)
	PDBVersionVC70    PDBVersion = 20000404 // VC 7.0 (2000-04-04)
	PDBVersionVC70Dep PDBVersion = 19990604 // VC 7.0 - deprecated (1999-06-04)
	PDBVersionVC80    PDBVersion = 20030901 // VC 8.0 (2003-09-01)
	PDBVersionVC110   PDBVersion = 20091201 // VC 11.0 (2009-12-01)
	PDBVersionVC140   PDBVersion = 20140508 // VC 14.0 (2014-05-08)
)

// GUID is a globally unique identifier.
type GUID [2]uint64

// parsePDBStreamHeader parses the given PDB stream header.
func (file *File) parsePDBStreamHeader(r io.Reader) (*PDBStreamHeader, error) {
	// Version.
	hdr := &PDBStreamHeader{}
	if err := binary.Read(r, binary.LittleEndian, &hdr.Version); err != nil {
		return nil, errors.WithStack(err)
	}
	// Date.
	var rawDate uint32
	if err := binary.Read(r, binary.LittleEndian, &rawDate); err != nil {
		return nil, errors.WithStack(err)
	}
	hdr.Date = time.Unix(int64(rawDate), 0)
	// Age.
	if err := binary.Read(r, binary.LittleEndian, &hdr.Age); err != nil {
		return nil, errors.WithStack(err)
	}
	// UniqueID.
	if err := binary.Read(r, binary.LittleEndian, &hdr.UniqueID); err != nil {
		return nil, errors.WithStack(err)
	}
	return hdr, nil
}

// StreamNameMap maps from stream name to stream number.
//
// ref: https://llvm.org/docs/PDB/PdbStream.html#named-stream-map
type StreamNameMap struct {
	// TODO: implement when we find a sample PDB that uses this.
}

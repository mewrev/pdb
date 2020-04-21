// Package pdb provides access to the PDB (Microsoft C/C++ program database)
// file format.
//
// ref: https://www.nationalarchives.gov.uk/pronom/fmt/1078
// ref: https://github.com/Microsoft/microsoft-pdb/blob/master/PDB/msf/msf.cpp
// ref: https://llvm.org/docs/PDB/MsfFile.html
package pdb

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"math"

	"github.com/pkg/errors"
)

// From https://github.com/microsoft/microsoft-pdb
//
//    +============+==============================+=====================================================================+
//    | Stream no. | Contents                     | Short Description                                                   |
//    +============+==============================+=====================================================================+
//    | 1          | Pdb (header)                 | Version information, and information to connect this PDB to the EXE |
//    | 2          | Tpi (Type manager)           | All the types used in the executable.                               |
//    | 3          | Dbi (Debug information)      | Holds section contributions, and list of ‘Mods’                     |
//    | 4          | NameMap                      | Holds a hashed string table                                         |
//    | 4-(n+4)    | n Mod’s (Module information) | Each Mod stream holds symbols and line numbers for one compiland    |
//    | n+4        | Global symbol hash           | An index that allows searching in global symbols by name            |
//    | n+5        | Public symbol hash           | An index that allows searching in public symbols by addresses       |
//    | n+6        | Symbol records               | Actual symbol records of global and public symbols                  |
//    | n+7        | Type hash                    | Hash used by the TPI stream.                                        |
//    +------------+------------------------------+---------------------------------------------------------------------+

// File is a PDB file.
type File struct {
	// File header of MSF.
	FileHdr *MSFHeader
	// Free page map.
	FreePageMap *FreePageMap
	// Stream table.
	StreamTbl *StreamTable

	// Contents of underlying PDB file.
	Data []byte // TODO: rename to buf
}

// ParseFile parses the given PDB file, reading from pdbPath.
func ParseFile(pdbPath string) (*File, error) {
	// Read PDB file contents.
	buf, err := ioutil.ReadFile(pdbPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	file := &File{
		Data: buf,
	}
	// Parse MSF file header.
	msfHdr, err := parseMSFHeader(bytes.NewReader(file.Data))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	file.FileHdr = msfHdr
	// TODO: parse FreePageMap.
	// Parse stream table.
	streamTblData := file.readStreamTable()
	streamTbl, err := file.parseStreamTable(bytes.NewReader(streamTblData))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	file.StreamTbl = streamTbl
	// TODO: parse streams.
	return file, nil
}

// readPage returns the contents of the given page.
func (file *File) readPage(pageNum int) []byte {
	pageSize := int(file.FileHdr.PageSize)
	start := pageNum * pageSize
	end := start + pageSize
	return file.Data[start:end]
}

// MSF signatures.
const (
	msfSignature = "Microsoft C/C++ program database 2.00\r\n\x1a\x4a\x47\x00\x00"
	// TODO: define signature for MSFBig.
)

// MSFHeader is the header of a multistream file (MSF). The MSF header is always
// at page 0.
//
// ref: https://llvm.org/docs/PDB/MsfFile.html#the-superblock
// ref: MSF_HDR
type MSFHeader struct {
	// File format identifier.
	Magic [44]byte
	// Page size in bytes.
	PageSize int32
	// Page number of free page map.
	FreePageMapPageNum uint16
	// Number of pages.
	NPages uint16
	// Stream information about the stream table.
	StreamTblInfo StreamInfo
	// Maps from stream page number to page number.
	PageNumMap []uint16 // length: math.Ceil(msfHdr.StreamTblInfo.Size / msfHdr.PageSize)
	// align until page boundry.
}

// parseMSFHeader parses the given MSF file header, reading from r.
func parseMSFHeader(r io.Reader) (*MSFHeader, error) {
	// Magic.
	msfHdr := &MSFHeader{}
	if err := binary.Read(r, binary.LittleEndian, &msfHdr.Magic); err != nil {
		return nil, errors.WithStack(err)
	}
	magic := string(msfHdr.Magic[:])
	if magic != msfSignature {
		return nil, errors.Errorf("invalid MSF signature; expected %q, got %q", msfSignature, magic)
	}
	// PageSize.
	if err := binary.Read(r, binary.LittleEndian, &msfHdr.PageSize); err != nil {
		return nil, errors.WithStack(err)
	}
	// FreePageMapPageNum.
	if err := binary.Read(r, binary.LittleEndian, &msfHdr.FreePageMapPageNum); err != nil {
		return nil, errors.WithStack(err)
	}
	// NPages.
	if err := binary.Read(r, binary.LittleEndian, &msfHdr.NPages); err != nil {
		return nil, errors.WithStack(err)
	}
	// StreamTblInfo.
	if err := binary.Read(r, binary.LittleEndian, &msfHdr.StreamTblInfo); err != nil {
		return nil, errors.WithStack(err)
	}
	// PageNumMap.
	streamTblNPages := int(math.Ceil(float64(msfHdr.StreamTblInfo.Size) / float64(msfHdr.PageSize))) // number of pages used by stream table.
	msfHdr.PageNumMap = make([]uint16, streamTblNPages)
	if err := binary.Read(r, binary.LittleEndian, &msfHdr.PageNumMap); err != nil {
		return nil, errors.WithStack(err)
	}
	// TODO: validate alignment until page boundry to be all zero?
	return msfHdr, nil
}

// StreamInfo specifies stream information.
//
// ref: SI_PERSIST
type StreamInfo struct {
	// Size in bytes of stream table.
	Size int32
	// ref: SI_PERSIST.mpspnpn
	Unknown int32
}

// FreePageMap specifies what pages are used/unused.
//
// ref: https://llvm.org/docs/PDB/MsfFile.html#the-free-block-map
// ref: FPM
type FreePageMap struct {
	// Each bit specifies whether the corresponding page is used or unused.
	//
	//    0 = used
	//    1 = unused
	PageBits []byte // length: msfHdr.PageSize
}

// StreamTable contains information about each stream of the MSF.
//
// Example [1]: Suppose a hypothetical PDB file with a 4KiB block size, and 4
// streams of lengths {1000 bytes, 8000 bytes, 16000 bytes, 9000 bytes}.
//
//    * Stream 0: ceil(1000 / 4096) = 1 block
//    * Stream 1: ceil(8000 / 4096) = 2 blocks
//    * Stream 2: ceil(16000 / 4096) = 4 blocks
//    * Stream 3: ceil(9000 / 4096) = 3 blocks
//
//    type stream_table struct {
//      nstreams = uint32(4)
//      stream_infos = []stream_info{{size: 1000}, {size: 8000}, {size: 16000}, {size: 9000}}
//      page_num_maps = [][]uint16{
//        {4},
//        {5, 6},
//        {11, 9, 7, 8},
//        {10, 15, 12},
//      },
//    }
//
// ref [1]: https://llvm.org/docs/PDB/MsfFile.html#the-stream-directory
// ref: StrmTbl
type StreamTable struct {
	// Number of streams.
	NStreams uint32
	// Stream information about each stream of the MSF.
	StreamInfos []StreamInfo // length: NStreams
	// Maps from stream number and stream page number to page number. Note that
	// the array is jagged, and as such, the length of the page number slices may
	// differ.
	PageNumMaps [][]uint16 // length of PageNumMaps[i]: math.Ceil(streamTbl.StreamInfos[i].Size / msfHdr.PageSize)
}

// readStreamTable reads the contents of the stream table, concatenating its
// pages together.
func (file *File) readStreamTable() []byte {
	streamTblNPages := int(math.Ceil(float64(file.FileHdr.StreamTblInfo.Size) / float64(file.FileHdr.PageSize))) // number of pages used by stream table.
	var streamTblData []byte
	for streamPageNum := 0; streamPageNum < streamTblNPages; streamPageNum++ {
		pageNum := int(file.FileHdr.PageNumMap[streamPageNum])
		pageData := file.readPage(pageNum)
		streamTblData = append(streamTblData, pageData...)
	}
	return streamTblData
}

// parseStreamTable parses the given stream table, reading from r.
func (file *File) parseStreamTable(r io.Reader) (*StreamTable, error) {
	// NStreams.
	streamTbl := &StreamTable{}
	if err := binary.Read(r, binary.LittleEndian, &streamTbl.NStreams); err != nil {
		return nil, errors.WithStack(err)
	}
	// StreamInfos.
	streamTbl.StreamInfos = make([]StreamInfo, streamTbl.NStreams)
	if err := binary.Read(r, binary.LittleEndian, &streamTbl.StreamInfos); err != nil {
		return nil, errors.WithStack(err)
	}
	// PageNumMaps.
	streamTbl.PageNumMaps = make([][]uint16, streamTbl.NStreams)
	for i := range streamTbl.PageNumMaps {
		streamNPages := int(math.Ceil(float64(streamTbl.StreamInfos[i].Size) / float64(file.FileHdr.PageSize)))
		streamTbl.PageNumMaps[i] = make([]uint16, streamNPages)
		if err := binary.Read(r, binary.LittleEndian, &streamTbl.PageNumMaps[i]); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return streamTbl, nil
}

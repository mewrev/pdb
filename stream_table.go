package pdb

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/pkg/errors"
)

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
//    type StreamTable struct {
//       NStreams = uint32(4)
//       StreamInfos = []StreamInfo{{Size: 1000}, {Size: 8000}, {Size: 16000}, {Size: 9000}}
//       PageNumMaps = [][]uint16{
//          {4},
//          {5, 6},
//          {11, 9, 7, 8},
//          {10, 15, 12},
//       },
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

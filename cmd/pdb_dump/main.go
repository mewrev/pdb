package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kr/pretty"
	"github.com/mewkiz/pkg/term"
	"github.com/mewrev/pdb"
	"github.com/pkg/errors"
)

var (
	// dbg is a logger with the "pdb_dump:" prefix which logs debug messages to
	// standard error.
	dbg = log.New(os.Stderr, term.CyanBold("pdb_dump:")+" ", 0)
	// warn is a logger with the "pdb_dump:" prefix which logs warning messages to
	// standard error.
	warn = log.New(os.Stderr, term.RedBold("pdb_dump:")+" ", 0)
)

func main() {
	flag.Parse()
	for _, pdbPath := range flag.Args() {
		if err := pdbDump(pdbPath); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

// pdbDump dumps the contents of the given PDB file.
func pdbDump(pdbPath string) error {
	file, err := pdb.ParseFile(pdbPath)
	if err != nil {
		return errors.WithStack(err)
	}
	file.Data = nil // TODO: remove
	/*
		for pageNum := 0; pageNum < int(file.FileHdr.NPages); pageNum++ {
			fmt.Printf("pageNum: %d, free: %v\n", pageNum, file.FreePageMap.IsFree(pageNum))
		}
	*/
	pretty.Println(file)
	fmt.Println()
	for streamNum, stream := range file.Streams {
		streamID := pdb.StreamID(streamNum)
		fmt.Printf("=== [ %v ] ===================================\n", streamID)
		fmt.Println()
		pretty.Println(stream)
		fmt.Println()
		switch stream := stream.(type) {
		case *pdb.StreamTable:
			// nothing to do.
		case *pdb.PDBStream:
			fmt.Println(streamID)
			fmt.Println("   Version:", stream.Hdr.Version)
			fmt.Println("   Date:", stream.Hdr.Date)
			fmt.Println("   Age:", stream.Hdr.Age)
			fmt.Println("   UniqueID:", stream.Hdr.UniqueID)
			fmt.Println()
		default:
			warn.Printf("not yet pretty-printing stream %T", stream)
		}
	}
	return nil
}

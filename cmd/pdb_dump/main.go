package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kr/pretty"
	"github.com/mewrev/pdb"
	"github.com/pkg/errors"
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
	for pageNum := 0; pageNum < int(file.FileHdr.NPages); pageNum++ {
		fmt.Printf("pageNum: %d, free: %v\n", pageNum, file.FreePageMap.IsFree(pageNum))
	}
	pretty.Println(file)
	return nil
}

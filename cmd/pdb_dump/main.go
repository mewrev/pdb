package main

import (
	"flag"
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
	pretty.Println(file)
	return nil
}

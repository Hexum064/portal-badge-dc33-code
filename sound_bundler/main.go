package main

import (
	"flag"
	"fmt"
)

type Args struct {
	SrcDir     string
	DestPath   string
	MapOutPath string
	MapInPath  string
}

func main() {

	args := Args{}

	flag.StringVar(&args.SrcDir, "src", ".", "Source directory of sound files")
	flag.StringVar(&args.DestPath, "dest", "sounds.bin", "Path of output bin file")
	flag.StringVar(&args.MapOutPath, "map-out", "map.csv", "Path of the output map csv file")
	flag.StringVar(&args.MapInPath, "map-in", "", "Path of the input map csv file")

	flag.Parse()

	fmt.Printf("Flags: %+v", args)

	//TODO: Check flags for errors
	//TODO: Output progress to console at each step
	//TODO: Output report of results (maybe with time)
	//TODO: Load the map file, if there is one, to get the order of the files
	//TODO: Get a list of all .wav files, recursively, from the source dir
	//TODO: Calculate size, in bytes, of LUT: 32 bit uints,
	// 1 int for number of entries, each entry will be 2 ints: 1 int for
	// start offset (in bytes), 1 int for size (in bytes)
	// Use a struct and create slice. Update the slice using indexes
	// Offsets all need to include size of LUT in bytes + 4 bytes for number of entries
	//TODO: Read each file and store the data in a large byte array, then update
	// the LUT with the number of bytes and offset. Note: if there is an input map,
	// order the files using that first, with all other files found appended to the end
	//TODO: Record the order in which the files are read into the map, using file names
	//TODO: Write number of files, LUT, and large byte array to output file
	//TODO: Write map csv, if specified

}

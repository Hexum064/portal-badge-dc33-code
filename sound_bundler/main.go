package main

import (
	"encoding/binary"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Terminal control codes
const (
	ShowCursor = "\033[?25h"
	HideCursor = "\033[?25l"
	// CursorUpFormat Requires formatting with number of lines
	CursorUpFormat = "\033[%dA"
	ClearLine      = "\r\033[K"
	KeyUp          = byte(65)
	KeyDown        = byte(66)
	KeyEscape      = byte(27)
	KeyEnter       = byte(13)
)

type Args struct {
	SrcDir     string
	DestPath   string
	MapOutPath string
	MapInPath  string
}

type DataBlock struct {
	Start uint32
	Len   uint32
}

func main() {

	args := Args{}

	// var fMap []string
	err := args.GetStartupFlags()

	if err != nil {
		log.Fatal(err)
		panic("could not load startup flags")
	}

	// if len(args.MapInPath) > 0 {

	// 	fMap, err = LoadMap(args.MapInPath)

	// 	if err != nil {
	// 		log.Fatal(err)
	// 		panic("could not load startup flags")
	// 	}
	// }

	fList, err := GetFileList(args.SrcDir)

	if err != nil {
		panic("could not get source files")
	}

	fListLen := len(fList)
	fValidList := make([]string, 0, fListLen)
	blocks := make([]DataBlock, 0, fListLen)
	fmt.Printf("Files found: %d. Validating...\n", fListLen)

	for i, f := range fList {

		fmt.Printf("Validating %d of %d\n", i+1, fListLen)

		valid, err := IsCorrectFormat(f)
		if err != nil {
			fmt.Printf("Error while checking file format: %v\n", err)
		}

		if !valid {
			fmt.Printf("File, %s, is not valid\n", f)
		}

		start, len, err := GetStartAndLen(f)

		if err != nil {
			fmt.Printf("Error while getting start and len: %v\n", err)
		}

		blocks = append(blocks, DataBlock{start, len})

		fValidList = append(fValidList, f)

		fmt.Print(HideCursor)
		fmt.Printf("\033[%dA", 1)
		fmt.Print(ShowCursor)
	}

	// fmt.Printf("%v\n", fValidList)

	fListLen = len(fValidList)
	fmt.Printf("Valid files found: %d. Collecting file data.\n", fListLen)

	dataStart := uint32(4 + (fListLen * 8))

	indexes := make([]byte, 0, (fListLen * 8))
	var data []byte

	for i, f := range fValidList {

		fmt.Printf("Processing %d of %d\n", i+1, fListLen)

		indexes = binary.LittleEndian.AppendUint32(indexes, dataStart)
		indexes = binary.LittleEndian.AppendUint32(indexes, blocks[i].Len)

		dataStart += blocks[0].Len

		dataBuff, err := GetDataBytes(f, blocks[i].Start, blocks[i].Len)

		if err != nil {
			panic(fmt.Errorf("error reading data bytes: %w", err))
		}

		data = append(data, dataBuff...)

		fmt.Print(HideCursor)
		fmt.Printf("\033[%dA", 1)
		fmt.Print(ShowCursor)
	}

	fmt.Printf("Done. Saving to %s\n", args.DestPath)
	SaveData(args.DestPath, uint32(fListLen), indexes, data)
	fmt.Printf("Writing file list to %s\n", args.MapOutPath)
	SaveFileList(args.MapOutPath, fValidList, true, true)
	fmt.Print("Done\n")

	// for i := 0; i < 10; i++ {
	// 	fmt.Println(i)
	// 	time.Sleep(1 * time.Second)
	// 	fmt.Printf(HideCursor)
	// 	fmt.Printf("\033[%dA", 1)
	// 	fmt.Printf(ShowCursor)
	// }

	//TODO: Check flags for errors
	//TODO: Output progress to console at each step
	//TODO: Output report of results (maybe with time)
	//TODO: Load the map file, if there is one, to get the order of the files
	//TODO: Get a list of all .wav files, recursively, from the source dir
	//TODO: Calculate size, in bytes, of LUT: 32 bit u_ints,
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

func (a *Args) GetStartupFlags() error {
	flag.StringVar(&a.SrcDir, "src", ".", "Source directory of sound files")
	flag.StringVar(&a.DestPath, "dest", "sounds.bin", "Path of output bin file")
	flag.StringVar(&a.MapOutPath, "map-out", "map.csv", "Path of the output map csv file")
	flag.StringVar(&a.MapInPath, "map-in", "", "Path of the input map csv file")

	flag.Parse()

	return nil
}

func LoadMap(path string) ([]string, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	r := csv.NewReader(f)
	recs, err := r.ReadAll()

	if err != nil {
		return nil, err
	}

	return recs[0], nil

}

func GetFileList(root string) ([]string, error) {

	var f []string = []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("[GetFileList] error reading file %s\n%w", path, err)
		}
		if !info.IsDir() {
			f = append(f, path)
		}
		// fmt.Println(path)
		return nil
	})

	//Don't need to add anything to the error
	if err != nil {
		return nil, err
	}

	return f, nil
}

// func UpdateFileOrder(allFiles []string, fMap []string)

func CheckStringBitField(offset int, value string, file *os.File) bool {
	buff := make([]byte, 4)
	_, err := file.ReadAt(buff, int64(offset))

	if err != nil {
		fmt.Printf("[CheckStringBitField] could not read file bites %d-%d: %v\n", offset, offset+3, err)
		return false
	}

	if string(buff) != value {
		fmt.Printf("[CheckStringBitField] Not a wave file. Did not find '%s' at %d\n", value, offset)
		return false
	}

	return true
}

func CheckIntBitField(offset int, len int, value uint, file *os.File) bool {
	buff := make([]byte, len)
	_, err := file.ReadAt(buff, int64(offset))

	if err != nil {
		fmt.Printf("[CheckIntBitField] could not read file bites %d-%d: %v\n", offset, offset+3, err)
		return false
	}

	var rVal uint = 0

	if len == 2 {
		rVal = uint(binary.LittleEndian.Uint16(buff))
	} else {
		rVal = uint(binary.LittleEndian.Uint32(buff))
	}

	if rVal != value {
		fmt.Printf("[CheckIntBitField] Did not find %d at %d\n", value, offset)
		return false
	}

	return true
}

func IsCorrectFormat(path string) (bool, error) {

	file, err := os.Open(path)

	if err != nil {
		return false, fmt.Errorf("[IsCorrectFormat] could not open file %s\n%w", path, err)
	}

	defer file.Close()

	if !(CheckStringBitField(0, "RIFF", file) &&
		CheckStringBitField(8, "WAVE", file) &&
		CheckIntBitField(20, 2, 1, file) &&
		CheckIntBitField(22, 2, 1, file) &&
		CheckIntBitField(24, 4, 22050, file) &&
		CheckIntBitField(34, 2, 16, file)) {
		return false, nil
	}

	return true, nil
}

func GetStartAndLen(path string) (start uint32, len uint32, err error) {

	file, err := os.Open(path)

	if err != nil {
		return 0, 0, fmt.Errorf("[GetStartAndLen] could not open file %s\n%w", path, err)
	}

	defer file.Close()

	buff := make([]byte, 4)
	offset := int64(36)

	_, err = file.ReadAt(buff, offset)

	if err != nil {
		return 0, 0, fmt.Errorf("[GetStartAndLen] could not read file bytes %s\n%w", path, err)
	}

	offset += 4

	if string(buff) == "LIST" {

		_, err = file.ReadAt(buff, offset)

		if err != nil {
			return 0, 0, fmt.Errorf("[GetStartAndLen] could not read file bytes %s\n%w", path, err)
		}

		//offset is 40 + 4 bytes for len + len
		offset += 4 + int64(binary.LittleEndian.Uint32(buff))

		_, err = file.ReadAt(buff, offset)

		if err != nil {
			return 0, 0, fmt.Errorf("[GetStartAndLen] could not read file bytes %s\n%w", path, err)
		}
		offset += 4
	}

	if string(buff) != "data" {
		return 0, 0, fmt.Errorf("[GetStartAndLen] unrecognized chunk (not data or LIST) %s", path)
	}

	_, err = file.ReadAt(buff, offset)

	if err != nil {
		return 0, 0, fmt.Errorf("[GetStartAndLen] could not read file bytes %s\n%w", path, err)
	}

	offset += 4

	return uint32(offset), binary.LittleEndian.Uint32(buff), nil
}

func GetDataBytes(path string, start uint32, len uint32) (data []byte, err error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("[GetDataBytes] could not open file %s\n%w", path, err)
	}

	defer file.Close()

	buff := make([]byte, len)
	_, err = file.ReadAt(buff, int64(start))

	if err != nil {
		return nil, fmt.Errorf("[GetDataBytes] could not read file bytes %s\n%w", path, err)
	}

	return buff, nil
}

func SaveData(path string, fileCnt uint32, indexes []byte, data []byte) error {
	// fmt.Printf("Indexes: %v\n", indexes)
	// fmt.Printf("Data: %v\n", data)
	allData := make([]byte, 0, 4+len(indexes)+len(data))
	allData = binary.LittleEndian.AppendUint32(allData, fileCnt)
	allData = append(allData, indexes...)
	allData = append(allData, data...)

	// fmt.Printf("allData: %v\n", allData)
	file, err := os.Create(path)

	if err != nil {
		return fmt.Errorf("[SaveData] could not create file %s\n%w", path, err)
	}

	defer file.Close()
	_, err = file.Write(allData)

	if err != nil {
		return fmt.Errorf("[SaveData] could not write data to file %s\n%w", path, err)
	}

	return nil

}

func SaveFileList(path string, files []string, withNewLine bool, withIndexNum bool) error {
	file, err := os.Create(path)
	output := ""
	if err != nil {
		return fmt.Errorf("[SaveFileList] could not create file %s\n%w", path, err)
	}

	defer file.Close()

	for i, f := range files {

		if withIndexNum {
			output = fmt.Sprintf("%04d,%s", i, f)
		} else {
			output = f
		}

		if withNewLine {
			output += "\n"
		} else {
			output += ","
		}

		_, err = file.WriteString(output)

		if err != nil {
			return fmt.Errorf("[SaveFileList] could not write to file %s\n%w", path, err)
		}
	}

	return nil
}

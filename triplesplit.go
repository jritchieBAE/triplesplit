package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main2() {

	fileName := flag.String("file", "", "path to input file")
	files := flag.Int("split", 2, "number of files to split across")
	upload := flag.Bool("upload", false, "upload files to Fuseki")

	flag.Parse()

	if *fileName == "" {
		fmt.Println("You have not specified an input file.")
		return
	}
	run(*fileName, *files, *upload)
}

func run(fileName string, files int, upload bool) error {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	prefix, extn := getParts(fileName)
	reader := bufio.NewScanner(f)
	var writer *destinationArray
	if upload {
		writer = newGraphArray(reader, prefix, extn, files)
	} else {
		writer = newFileArray(reader, prefix, extn, files)
	}
	defer writer.close()
	err = writer.doleOut()
	if err != nil {
		return err
	}
	fmt.Println(writer.Success())
	return nil
}

func getParts(name string) (string, string) {
	paths := strings.Split(name, "/")
	fileName := paths[len(paths)-1]
	parts := strings.Split(fileName, ".")
	extension := parts[len(parts)-1]
	name = strings.Join(parts[0:len(parts)-1], "")
	return name, extension
}

func newFileArray(reader *bufio.Scanner, prefix, extn string, count int) *destinationArray {
	writer := &destinationArray{
		source:     reader,
		namePrefix: prefix,
		nameSuffix: extn,
		count:      int64(count),
	}
	destinationCreator := func() (destinationWriter, error) {
		f, err := os.Create(writer.filename())
		if err != nil {
			return nil, err
		}
		return &fileWriter{f}, nil
	}
	writer.destinationCreator = destinationCreator
	writer.Success = func() string {
		return fmt.Sprintf("%v lines written into %v files.", writer.total, math.Min(float64(writer.count), float64(writer.total)))
	}
	return writer
}

func newGraphArray(reader *bufio.Scanner, prefix, extn string, count int) *destinationArray {
	writer := newFileArray(reader, prefix, extn, count)
	writer.Success = func() string {

		fmt.Println("Uploading...")
		start := time.Now()
		for writer.destinationIndex = 0; writer.destinationIndex < int64(count); writer.destinationIndex++ {
			url := fmt.Sprintf("%s?graph=%s", `http://localhost:3030/test`, writer.indexAsBinary())
			file := writer.filename()
			r, err := os.Open(file)
			if err != nil {
				return err.Error()
			}
			_, err = http.Post(url, "application/n-triples", r)
			if err != nil {
				return err.Error()
			}
		}
		total := time.Since(start)
		return fmt.Sprintf("%v lines written into %v files and uploaded to Fuseki in %v", writer.total, math.Min(float64(writer.count), float64(writer.total)), total)
	}
	return writer
}

type destinationArray struct {
	source             *bufio.Scanner
	namePrefix         string
	nameSuffix         string
	count              int64
	total              int64
	destinationCreator func() (destinationWriter, error)
	destinationIndex   int64
	destinations       []destinationWriter
	Success            func() string
}

func (arr *destinationArray) doleOut() error {
	for arr.source.Scan() {
		line := arr.source.Text()
		if line == "" {
			return nil
		} else {
			line = fmt.Sprintf("%s\n", line)
			err := arr.write(strings.NewReader(line))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (arr *destinationArray) write(content io.Reader) error {
	if arr.destinations == nil {
		arr.destinations = make([]destinationWriter, arr.count)
	}
	if arr.destinations[arr.destinationIndex] == nil {
		dest, err := arr.destinationCreator()
		if err != nil {
			return err
		}
		arr.destinations[arr.destinationIndex] = dest
	}
	dest := arr.destinations[arr.destinationIndex]
	err := dest.Write(content)
	if err == nil {
		arr.nextFile()
	}
	arr.total++
	return err
}

func (arr *destinationArray) filename() string {
	name := arr.indexAsBinary()
	return fmt.Sprintf("%s-%s.%s", arr.namePrefix, name, arr.nameSuffix)
}

func (arr *destinationArray) indexAsBinary() string {
	name := strconv.FormatInt(arr.destinationIndex+1, 2)
	highestNameLength := len(strconv.FormatInt(arr.count, 2))
	if len(name) < highestNameLength {
		zeros := strings.Repeat("0", highestNameLength-len(name))
		name = zeros + name
	}
	return name
}

func (arr *destinationArray) nextFile() {
	arr.destinationIndex++
	if arr.destinationIndex >= arr.count {
		arr.destinationIndex = 0
	}
}

func (arr *destinationArray) close() {
	for _, file := range arr.destinations {
		file.Close()
	}
}

type destinationWriter interface {
	Write(io.Reader) error
	Close()
}

type fileWriter struct {
	file *os.File
}

func (fw *fileWriter) Close() {
	fw.file.Close()
}

func (fw *fileWriter) Write(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	_, err = fw.file.WriteString(string(b))
	return err
}

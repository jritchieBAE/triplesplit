package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {

	fileName := flag.String("file", "", "path to input file")
	files := flag.Int("split", 2, "number of files to split across")

	flag.Parse()

	if *fileName == "" {
		fmt.Println("You have not specified an input file.")
		return
	}

	f, err := os.Open(*fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	prefix, extn := getParts(*fileName)
	reader := bufio.NewScanner(f)
	writer := &fileArray{
		source:     reader,
		namePrefix: prefix,
		nameSuffix: extn,
		count:      int64(*files),
	}
	defer writer.close()
	err = writer.doleOut()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v lines written into %v files.\n", writer.total, math.Min(float64(writer.count), float64(writer.total)))
	}
}

func getParts(name string) (string, string) {
	paths := strings.Split(name, "/")
	fileName := paths[len(paths)-1]
	parts := strings.Split(fileName, ".")
	extension := parts[len(parts)-1]
	name = strings.Join(parts[0:len(parts)-1], "")
	return name, extension
}

type fileArray struct {
	source     *bufio.Scanner
	namePrefix string
	nameSuffix string
	count      int64
	total      int
	file       int64
	files      []*os.File
}

func (fa *fileArray) doleOut() error {
	for fa.source.Scan() {
		line := fa.source.Text()
		if line == "" {
			return nil
		} else {
			err := fa.write(fmt.Sprintf("%v\n", line))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fa *fileArray) write(content string) error {
	if fa.files == nil {
		fa.files = make([]*os.File, fa.count)
	}
	if fa.files[fa.file] == nil {
		f, err := os.Create(fa.filename())
		if err != nil {
			return err
		}
		fa.files[fa.file] = f
	}
	file := fa.files[fa.file]
	_, err := file.WriteString(content)
	if err == nil {
		fa.nextFile()
	}
	fa.total++
	return err
}

func (fa *fileArray) filename() string {
	highestNameLength := len(strconv.FormatInt(fa.count, 2))
	name := strconv.FormatInt(fa.file+1, 2)
	if len(name) < highestNameLength {
		zeros := strings.Repeat("0", highestNameLength-len(name))
		name = zeros + name
	}
	return fmt.Sprintf("%s-%s.%s", fa.namePrefix, name, fa.nameSuffix)
}

func (fa *fileArray) nextFile() {
	fa.file++
	if fa.file >= fa.count {
		fa.file = 0
	}
}

func (fa *fileArray) close() {
	for _, file := range fa.files {
		file.Close()
	}
}

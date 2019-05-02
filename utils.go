package data

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"path"
	"strings"

	"github.com/gonum/matrix/mat64"
	. "github.com/nimezhu/netio"
)

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
func sprintMat64(t mat64.Matrix) string {
	r, c := t.Dims()
	var buffer bytes.Buffer
	for i := 0; i < r; i++ {
		buffer.WriteString(fmt.Sprintf("%.0f", t.At(i, 0)))
		for j := 1; j < c; j++ {
			buffer.WriteString(fmt.Sprintf("\t%.0f", t.At(i, j)))
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func sprintMat64_2(t mat64.Matrix) string {
	r, c := t.Dims()
	var buffer bytes.Buffer
	for i := 0; i < r; i++ {
		buffer.WriteString(fmt.Sprintf("%.2f", t.At(i, 0)))
		for j := 1; j < c; j++ {
			buffer.WriteString(fmt.Sprintf("\t%.2f", t.At(i, j)))
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func float64bytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func matrixToBytes(m mat64.Matrix) []byte {
	var s bytes.Buffer
	br := make([]byte, 4)
	bc := make([]byte, 4)

	r, c := m.Dims()
	binary.LittleEndian.PutUint32(br, uint32(r))
	binary.LittleEndian.PutUint32(bc, uint32(c))
	s.Write(br)
	s.Write(bc)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			f := m.At(i, j)
			//fmt.Println(i, j, f)
			s.Write(float64bytes(f))
		}
	}
	return s.Bytes()
}
func bytesToMatrix(b []byte) mat64.Matrix { //todo
	reader := bytes.NewReader(b)
	r32, _ := ReadInt(reader)
	c32, _ := ReadInt(reader)
	r := int(r32)
	c := int(c32)
	mat := make([]float64, r*c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			mat[i*r+j], _ = ReadFloat64(reader)
		}
	}
	matrix := mat64.NewDense(r, c, mat)
	fmt.Println(sprintMat64(matrix))
	return matrix
}

func loadURI(uri string) map[string]string {
	ext := strings.ToLower(path.Ext(uri))
	uriMap := make(map[string]string)
	if ext == ".txt" || ext == ".tsv" {
		reader, err := NewReadSeeker(uri)
		checkErr(err)
		r := csv.NewReader(reader)
		r.Comma = '\t'
		r.Comment = '#'
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			log.Println(record)
			uriMap[record[0]] = record[1]
		}
	} else {
		if len(uri) > 0 {
			uriMap["default"] = uri
		}
	}
	return uriMap
}

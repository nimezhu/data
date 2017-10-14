package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/gonum/matrix/mat64"
	. "github.com/nimezhu/netio"
)

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func Float64bytes(float float64) []byte {
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
			s.Write(Float64bytes(f))
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
	//fmt.Println(row, col)
}

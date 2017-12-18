package data

import (
	"bytes"
	"fmt"
	"log"

	"github.com/gonum/matrix/mat64"
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

package data

import (
	"strconv"
	"strings"
)

type Bed3 interface {
	Chr() string
	Start() int
	End() int
}
type Bed4 struct {
	chr   string
	start int
	end   int
	name  string
}

func (b Bed4) Chr() string {
	return b.chr
}

func (b Bed4) Start() int {
	return b.start
}
func (b Bed4) End() int {
	return b.end
}
func (b Bed4) Id() string {
	return b.name
}

type ShortBed interface {
	Chr() string
	Start() int
	End() int
}

func regionText(b ShortBed) string {
	s := strconv.Itoa(b.Start())
	e := strconv.Itoa(b.End())
	return b.Chr() + ":" + s + "-" + e
}

//TODO
func regionsText(bs []Bed4) string { //TODO interface array function
	r := make([]string, len(bs))
	for i, v := range bs {
		r[i] = regionText(v)
	}
	return strings.Join(r, ",")
}
func bedsText(bs []Bed3) string { //TODO interface array function
	r := make([]string, len(bs))
	for i, v := range bs {
		r[i] = regionText(v)
	}
	return strings.Join(r, ",")
}

package data

import (
	"errors"
	"log"
)

var (
	binOffsets    = []uint{512 + 64 + 8 + 1, 64 + 8 + 1, 8 + 1, 1, 0}
	binFirstShift = uint(17)
	binNextShift  = uint(3)
	binLength     = []uint{4096 + 512 + 64 + 8 + 1}
)

type RangeI interface {
	Start() int
	End() int
}
type NamedRangeI interface {
	RangeI
	Id() string
}
type BedI interface {
	RangeI
	Chr() string
}
type NamedBedI interface {
	BedI
	Id() string
}

func range2bin(start uint, end uint) uint {
	startBin := start
	endBin := end - 1
	startBin >>= binFirstShift
	endBin >>= binFirstShift
	for _, v := range binOffsets {
		if startBin == endBin {
			return v + startBin
		}
		startBin >>= binNextShift
		endBin >>= binNextShift

	}
	return 0
}
func iterRangeOverlapBins(start uint, end uint) <-chan uint {
	ch := make(chan uint)
	go func() {
		startBin := start
		endBin := end - 1
		startBin >>= binFirstShift
		endBin >>= binFirstShift
		for _, v := range binOffsets {
			for j := startBin; j < endBin+1; j++ {
				ch <- j + v
			}
			startBin >>= binNextShift
			endBin >>= binNextShift
		}
		close(ch)
	}()
	return ch
}

func bin2range(bin uint) (uint, uint) {
	binShift := binFirstShift
	for _, v := range binOffsets {
		if bin-v >= 0 {
			bin = bin - v
			break
		}
		binShift += binNextShift
	}
	return bin << binShift, (bin + 1) << binShift
}

func bin2length(bin uint) uint {
	start, end := bin2range(bin)
	return end - start
}

func bin2level(bin uint) int {
	for i, v := range binOffsets {
		if bin-v >= 0 {
			return 4 - i
		}
	}
	return 0
}

type BinIndex map[int][]NamedRangeI

func newBinIndex() BinIndex {
	return make(map[int][]NamedRangeI)
}

type BinIndexMap struct {
	Data map[string]BinIndex
}

func NewBinIndexMap() *BinIndexMap {
	return &BinIndexMap{make(map[string]BinIndex)}
}
func (c *BinIndexMap) Load(b []NamedBedI) error {
	var err error
	for _, v := range b {
		err = c.Insert(v)
		if err != nil {
			return err
		}
	}
	return err
}
func (c *BinIndexMap) Insert(b NamedBedI) error {
	chr := b.Chr()
	if _, ok := c.Data[chr]; !ok {
		c.Data[chr] = newBinIndex()
	}
	v, _ := c.Data[chr]
	bin := range2bin(uint(b.Start()), uint(b.End()))
	if _, ok := v[int(bin)]; !ok {
		v[int(bin)] = []NamedRangeI{}
	}
	v[int(bin)] = append(v[int(bin)], b)
	return nil
}
func overlap(a RangeI, b RangeI) bool {
	return a.Start() < b.End() && b.Start() < a.End()
}
func (d *BinIndexMap) QueryRegion(chr string, start int, end int) (<-chan NamedRangeI, error) {
	q := Bed4{chr, start, end, "noname"}
	return d.Query(q)
}
func (d *BinIndexMap) Query(b BedI) (<-chan NamedRangeI, error) { // need to reflect
	bedCh := make(chan NamedRangeI)
	chr, ok := d.Data[b.Chr()]
	if !ok {
		return nil, errors.New("chr not found")
	}
	//n = range2bin(uint(b.Start()), uint(b.End()))
	go func() {
		for bin := range iterRangeOverlapBins(uint(b.Start()), uint(b.End())) {
			if values, ok := chr[int(bin)]; ok {
				for _, v := range values {
					if overlap(v, b) {
						bedCh <- v
					}
				}
			}
		}
		close(bedCh)
	}()

	return bedCh, nil
}

/* TODO Delete */
func (d *BinIndexMap) Delete(b NamedBedI) error {
	chr, ok := d.Data[b.Chr()]
	if !ok {
		return errors.New("chr not found")
	}
	for bin := range iterRangeOverlapBins(uint(b.Start()), uint(b.End())) {
		if values, ok := chr[int(bin)]; ok {
			for i, v := range values {
				if v.Id() == b.Id() && v.Start() == b.Start() && v.End() == b.End() {
					values = append(values[:i], values[i+1:]...)
					log.Println("delete", i)
					chr[int(bin)] = values
					break
				}
			}
		}
	}
	return nil
}

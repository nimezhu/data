package data

import "testing"

func TestCytoBand(t *testing.T) {
	m := &CytoBandManager{"band", make(map[string]*CytoBand)}
	m.Add("hg19")
	t.Log(m.List())
}

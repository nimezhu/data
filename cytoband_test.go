package data

import "testing"

func TestCytoBand(t *testing.T) {
	//uri := "http://hgdownload.cse.ucsc.edu/goldenPath/hg19/database/cytoBand.txt.gz"
	m := &CytoBandManager{"band", make(map[string]*CytoBand)}
	m.Add("hg19")
	t.Log(m.List())
}

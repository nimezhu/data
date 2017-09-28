package data

import "testing"

/*
func TestGetSeq(t *testing.T) {
	//uri := "http://hgdownload.cse.ucsc.edu/goldenPath/hg19/database/cytoBand.txt.gz"
	s := NewSeqServer()

	for i := 1; i < 20; i++ {
		seq2, _ := s.GetSeq("mm10", "chr1", i*100000+1700000, i*100000+1701000)
		t.Log(seq2)
	}
}
*/
func TestStaticGetSeq(t *testing.T) {
	//uri := "http://hgdownload.cse.ucsc.edu/goldenPath/hg19/database/cytoBand.txt.gz"
	/*
		seq, _ := s.GetSeq("hg19", "chr1", 700000, 700100)
		t.Log(seq)
		seq1, _ := s.GetSeq("hg19", "chr1", 700100, 700200)
		t.Log(seq1)
	*/
	for i := 1; i < 2; i++ {
		seq2, _ := GetSeq("mm10", "chr1", i*100000+1700000, i*100000+1701000)
		t.Log(seq2)
	}
}

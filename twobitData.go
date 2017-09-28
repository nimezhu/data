package data

import (
	"errors"

	"github.com/aebruno/twobit"
	"github.com/nimezhu/netio"
)

var genomes = map[string]string{
	"hg19": "http://hgdownload.soe.ucsc.edu/goldenPath/hg19/bigZips/hg19.2bit",
	"hg38": "http://hgdownload.soe.ucsc.edu/goldenPath/hg38/bigZips/hg38.2bit",
	"mm9":  "http://hgdownload.soe.ucsc.edu/goldenPath/mm9/bigZips/mm9.2bit",
	"mm10": "http://hgdownload.soe.ucsc.edu/goldenPath/mm10/bigZips/mm10.2bit",
	//"mm10": "/home/zhuxp/Data/genome/mm10/mm10.2bit",
}

/*
type SeqServer struct {
	Readers map[string]*twobit.Reader
}

func NewSeqServer() SeqServer {
	return SeqServer{
		make(map[string]*twobit.Reader),
	}
}
func (s SeqServer) Open(genome string) error {
	uri, ok1 := genomes[genome]
	if !ok1 {
		return errors.New("genome not support")
	}
	if _, ok := s.Readers[genome]; !ok {
		f, err := netio.Open(uri)
		if err != nil {
			return err
		}
		s.Readers[genome], err = twobit.NewReader(f)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s SeqServer) GetSeq(genome string, chr string, start int, end int) (string, error) {
	if _, ok := genomes[genome]; ok {
		//defer tb.Close()
		if _, ok := s.Readers[genome]; !ok {
			err := s.Open(genome)
			if err != nil {
				return "", err
			}
		}
		q, err := s.Readers[genome].ReadRange(chr, start, end)
		if err != nil {
			//t.Errorf("Failed to read sequence: %s", err)
			return "", err
		}
		return string(q), err
	}
	return "", errors.New("not support query genome version")
}
*/
func GetSeq(genome string, chr string, start int, end int) (string, error) {
	if v, ok := genomes[genome]; ok {
		f, err := netio.Open(v)
		defer f.Close()
		if err != nil {
			return "", err
		}
		tb, err := twobit.NewReader(f)
		if err != nil {
			return "", err
		}
		//defer tb.Close()
		q, err := tb.ReadRange(chr, start, end)
		if err != nil {
			//t.Errorf("Failed to read sequence: %s", err)
			return "", err
		}
		return string(q), err
	}
	return "", errors.New("not support query genome version")
}

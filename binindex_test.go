package data

import "testing"

func TestBinIndex(t *testing.T) {
	b := []NamedBedI{
		Bed4{"chr1", 1, 30000, "probe1"},
		Bed4{"chr2", 2, 40000, "probe2"},
		Bed4{"chr2", 1, 30000, "probe3"},
		Bed4{"chr5", 1000000, 1300000, "probe4"},
	}
	bin := NewBinIndexMap()
	bin.Load(b)
	c := Bed4{"chr2", 300, 700000, "query"}
	if ch, err := bin.Query(c); err == nil {
		for v := range ch {
			t.Log(v)
		}
	}

}

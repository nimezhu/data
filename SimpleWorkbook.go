package data

type SimpleWorkbook struct { //string sheets
	SheetNames []string
	Sheets     map[string][][]string
}



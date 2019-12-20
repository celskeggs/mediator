package ast

type File struct {
	Definitions []Definition
	SearchPath  []string
	Maps        []string
}

func (f *File) Extend(file *File) {
	f.Definitions = append(f.Definitions, file.Definitions...)
	f.SearchPath = append(f.SearchPath, file.SearchPath...)
	f.Maps = append(f.Maps, file.Maps...)
}

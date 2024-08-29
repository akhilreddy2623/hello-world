package enums

type APIRequestType int

const (
	NoneAPIRequestType APIRequestType = iota
	Forte
	FilmAPIUploadFile
	FilmAPIGetFileLines
	FilmAPIGetFilesInFolder
	FilmAPIDeleteFile
	Workday
)

func (a APIRequestType) String() string {
	return [...]string{
		"",
		"forte",
		"filmapiuploadfile",
		"filmapigetfilelines",
		"filmapigetfilesinfolder",
		"filmapideletefile",
		"workday"}[a]
}

func (a APIRequestType) EnumIndex() int {
	return int(a)
}

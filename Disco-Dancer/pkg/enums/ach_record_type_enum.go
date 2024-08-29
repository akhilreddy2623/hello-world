package enums

type AchRecordType int

const (
	FileHeader   ACHAccountType = 1
	BatchHeader  AchRecordType  = 5
	EntryDetail  AchRecordType  = 6
	AdendaRecord AchRecordType  = 7
	BatchFooter  AchRecordType  = 8
	Filefooter   AchRecordType  = 9
)

func (c AchRecordType) EnumIndex() int {
	return int(c)
}

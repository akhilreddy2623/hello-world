package app

type FileRecord struct {
	FileId                int
	RecordId              int64
	RecordContent         string
	BusinessFileType      string
	TotalRecordCount      int
	FilePath              string
	ArchiveFolderLocation string
}

type FileRecordFeedback struct {
	FileId                int
	RecordId              int64
	BusinessFileType      string
	TotalRecordCount      int
	IsError               bool
	FilePath              string
	ArchiveFolderLocation string
}

type FileProcessingCompleted struct {
	BusinessFileType string
}

package inbound_fileprocessor

type FileProcessorInput struct {
	PollingDurationInMins  int
	FolderLocation         string
	BusinessFileType       string
	ArchiveFolderLocation  string
	ProcessRecordTopicName string
}

type Record struct {
	FileId                int
	RecordId              int64
	RecordContent         string
	BusinessFileType      string
	TotalRecordCount      int
	FilePath              string
	ArchiveFolderLocation string
}

type RecordFeedback struct {
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

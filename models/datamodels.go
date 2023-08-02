package models

type CommitModel struct {
	Id           string
	Author       string
	Committer    string
	CommitMsg    string
	ParentCommit *CommitModel
}

type Tree struct {
	FileName string
}

type Blob struct {
	Content  string
	FileName string
}

type FileMetaData struct {
	Path             string
	Status           FileStatus
	ModificationTime string
	GoToStaging      bool
}

type FileStatus string

var (
	StatusCreated FileStatus = "Created"
	StatusUpdated FileStatus = "Updated"
)

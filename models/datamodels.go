package models

type CommitModel struct {
	CommitMsg     string
	ParentCommit  *CommitModel
	CommitVersion int
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

type WritableServer struct {
	ListAddr    string
	Connections []string
}

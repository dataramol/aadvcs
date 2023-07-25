package models

type CommitModel struct {
	Id           string
	Author       string
	Committer    string
	CommitMsg    string
	ParentCommit *CommitModel
}

type Tree struct {
	FileMode string
	FileName string
}

type Blob struct {
	Content  string
	FileName string
}

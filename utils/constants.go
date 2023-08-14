package utils

import (
	"path/filepath"
)

var (
	StatusFile      = filepath.Join(AadvcsRootDirName, "status.txt")
	StagingAreaFile = filepath.Join(AadvcsRootDirName, "staging_area.txt")
)

const (
	AadvcsRootDirName           = ".aadvcs"
	AadvcsCommitDirPath         = ".aadvcs/commit"
	AadvcsCheckoutDirPath       = ".aadvcs/checkout"
	AadvcsStatusFilePath        = ".aadvcs/status.txt"
	AadvcsStagingFilePath       = ".aadvcs/staging_area.txt"
	AadvcsNetworkConfigFilePath = ".aadvcs/network.json"
	AadvcsTimeFormat            = "2006-01-02 03:04:05"
	AadvcsCommitMetadataFile    = "metadata.txt"
)

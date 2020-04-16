package lib

var AESKEY string

const (
	//keyprefix
	ChunkPrefix = "/chunk"
	FilePrefix  = "/path"
	//key
	AbsPathKey   = "abspathprefix"
	ChunkSizeKey = "sectorsize"
	TypeKey      = "type"
	EncTypeKey   = "enctype"
	SlicesKey    = "slices"

	DbType   = "db"
	MetaType = "meta"
	//state
	Rejected = "rejected"
	Accepted = "accepted"
	Started  = "started"
	Failed   = "failed"
	Staged   = "staged"
	Complete = "complete"

	TempFiledir = "/root/storagetemp"
)

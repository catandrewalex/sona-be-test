package util

type InsertionType string

const (
	InsertionType_New          InsertionType = "NEW"
	InsertionType_FromExisting InsertionType = "FROM_EXISTING"
)

func (t InsertionType) String() string {
	return string(t)
}

var ValidInsertionTypes = map[InsertionType]struct{}{
	InsertionType_New:          {},
	InsertionType_FromExisting: {},
}

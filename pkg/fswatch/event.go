package fswatch

const (
	ReadFile   = iota // File read event
	WriteFile         // File write event
	DeleteFile        // File deletion event
	CreateFile        // File creation event
	CreateDir         // Directory creation event
	DeleteDir         // Directory deletion event
)

// todo: add data field
type Event struct {
	Path string
	Type uint8
}

package drone

type Build struct {
	Message string
	Number  int64
	Before  string
	After   string
	Source  string // name of the branch
}

func (b *Build) GetMessage() string {
	return b.Message
}

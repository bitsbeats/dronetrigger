package drone

type Build struct {
	Message string
	Number  int64
	Before  string
	After   string
}

func (b *Build) GetMessage() string {
    return b.Message
}

package core

type (
	// Build is a Drone build
	Build struct {
		Message string
		Number  int64
		Before  string
		After   string
		Source  string
	}

	// Drone is a api client for Drone
	Drone interface {
		RebuildLastBuild(repo, ref string) (*Build, error)
	}
)

func (b *Build) GetMessage() string {
	return b.Message
}

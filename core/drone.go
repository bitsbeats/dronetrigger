package core

type (
	// Build is a Drone build
	Build struct {
		Message string `json:"message"`
		Number  int64  `json:"number"`
		Before  string `json:"before"`
		After   string `json:"after"`
		Source  string `json:"source"`
		Event   string `json:"event"`
	}

	// Drone is a api client for Drone
	Drone interface {
		RebuildLastBuild(repo, ref string) (*Build, error)
		RebuildLastTag(repo string) (*Build, error)
		PromoteLastBuild(repo, ref, target string) (*Build, error)
		PromoteLastTag(repo, target string) (*Build, error)
	}
)

func (b *Build) GetMessage() string {
	return b.Message
}

package build

type BuildLog interface {
	Log(msg string)
	Error(msg string)
}

package types

type IMsrv interface {
	Start() func(error)
	PathInit(path string) error
	PathFini(path string) error
}

package types

type ImSrv interface {
	Start() func(error)
	PathInit(path string) error
	PathFini(path string) error
}

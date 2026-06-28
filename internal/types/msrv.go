package types

type ImSrv interface {
	Start() func(error)
	PathInit(path string)
	PathFini(path string)
}

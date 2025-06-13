package app_io

type Pauser interface {
	Pause() error
	UnPause() error
}

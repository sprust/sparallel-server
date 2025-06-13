package rpc

import (
	"io"
	"sparallel_server/pkg/foundation/app_io"
)

type ServerInterface interface {
	io.Closer
	app_io.Pauser
}

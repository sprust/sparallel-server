package rpc_ping_pong

import (
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"net/rpc"
	"os"
	"testing"
	"time"
)

func init() {
	if err := godotenv.Load("../../../.env"); err != nil {
		panic(err)
	}

	// TODO: validate env variables
}

func TestPingPong_Ping(t *testing.T) {
	rpcPort := os.Getenv("RPC_PORT")

	client, err := rpc.Dial("tcp", ":"+rpcPort)

	assert.NoError(t, err)

	defer func(client *rpc.Client) {
		err := client.Close()

		assert.NoError(t, err)
	}(client)

	message := time.Now().String()

	args := PingArgs{
		Message: message,
	}

	var result PingResult

	err = client.Call("PingPongServer.Ping", args, &result)

	assert.NoError(t, err)

	assert.Equal(t, message, result.Message)
}

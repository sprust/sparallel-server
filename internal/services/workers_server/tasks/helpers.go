package tasks

import "time"

func isTimeout(unixTimeout int, headStart int) bool {
	now := time.Now().Unix()

	return (int64(unixTimeout) - now) < -int64(headStart)
}

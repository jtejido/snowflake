package snowflake

import "time"

func TilNextMillis(lastTimestamp int64) (timestamp int64) {

	for timestamp <= lastTimestamp {
	  timestamp = time.Now().UnixNano() / 1000000
	}
	return
}
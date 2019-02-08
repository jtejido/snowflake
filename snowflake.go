// Copyright 2010-2012 Twitter, Inc.
// license that can be found in the LICENSE file.

// Snowflake is a network service for generating unique ID numbers at high scale with some simple guarantees. It gives single and multi worker options.
package snowflake

import "time"

// TW_EPOCH is set to the twitter snowflake's epoch of Nov 04 2010 01:42:54 UTC
const TW_EPOCH int64 = 1288834974657

type Node interface {
	GetID(t time.Time) (ID, error)
	LastTimeStamp() time.Time
	NextID() (ID, error)
}

type ID interface {
	Int64() int64
	String() string
	Bytes() []byte
	Time() int64
	Node() int64
	DataCenter() int64
	Sequence() int64
}

// Copyright 2010-2012 Twitter, Inc.
// license that can be found in the LICENSE file.

// package multi follows twitter's original scala implementation at https://github.com/twitter-archive/snowflake
package multi

import (
	"github.com/jtejido/snowflake"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (

	nodeIdBits uint8 = 5
	maxNodeId   int64 = -1 ^ (-1 << nodeIdBits)
	datacenterIdBits uint8 = 5
	maxDatacenterId   int64 = -1 ^ (-1 << datacenterIdBits)
	sequenceBits uint8 = 12
	maxSequenceBits uint8 = -1 ^ (-1 << sequenceBits)

	nodeIdShift uint8 = sequenceBits
	nodeMask  int64 = maxNodeId << sequenceBits
	datacenterIdShift uint8 = sequenceBits + nodeIdBits
	datacenterMask  int64 = maxDatacenterId << (nodeIdBits + sequenceBits)
	timestampLeftShift uint8 = sequenceBits + nodeIdBits + datacenterIdBits

	sequenceMask int64 = -1 ^ (-1 << sequenceBits)
	lastTimeStamp int64 = -1
)

// A JSONSyntaxError is returned from UnmarshalJSON if an invalid ID is provided.
type JSONSyntaxError struct{ original []byte }

func (j JSONSyntaxError) Error() string {
	return fmt.Sprintf("invalid snowflake ID %q", string(j.original))
}

var (
	ErrInvalidNode = errors.New("node Id must be between 0 and " + strconv.FormatInt(maxNodeId, 10))
	ErrInvalidDataCenter = errors.New("datacenter Id must be between 0 and " + strconv.FormatInt(maxDatacenterId, 10))
	ErrForwardTime = errors.New("You wish to generate new ID. use NextID() instead.")
)

// A Node struct holds the basic information needed for a snowflake generator
// node
type multiWorker struct {
	mu   sync.RWMutex
	lastTimeStamp int64
	nodeID int64
	datacenterId int64
	sequence int64
}

type MultiWorkerID int64

// New returns a new IDWorker
func New(nodeId, datacenterId int64) (*multiWorker, error) {

	if (nodeId > maxNodeId || nodeId < 0) {
		return nil, ErrInvalidNode
	}

	if (datacenterId > maxDatacenterId || datacenterId < 0) {
		return nil, ErrInvalidDataCenter
	}

	return &multiWorker{
		lastTimeStamp: 0,
		nodeID: nodeId,
		datacenterId: datacenterId,
		sequence: 0,
	}, nil
}

// GetID gets the corresponding ID given a timestamp and sequence
func (n *multiWorker) GetID(t time.Time, s int64) (id MultiWorkerID, err error) {
 	
 	n.mu.Lock()

	now := t.UnixNano() / 1000000

	if n.lastTimeStamp != 0 && now > n.lastTimeStamp {
      	return MultiWorkerID(0), ErrForwardTime
    }

	// no prior knowledge of how many times an id has been generated for the same ms, store id.sequence() to be used here
    // TO-DO: get all IDs up to maxSequenceBits? (O(n) where n is maxSequenceBits)
	id = MultiWorkerID((now-snowflake.TW_EPOCH)<<timestampLeftShift | 
		(n.datacenterId << datacenterIdShift) |
		(n.nodeID << nodeIdShift) |
		s)


	n.mu.Unlock()

	return
}

// LastTimeStamp returns last time.Time
func (n multiWorker) LastTimeStamp() time.Time {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return time.Unix(0, n.lastTimeStamp * 1000000)
}

// NextID creates and returns a unique snowflake ID
func (n *multiWorker) NextID() (id MultiWorkerID, err error) {

	n.mu.Lock()

	now := time.Now().UnixNano() / 1000000

	if now < n.lastTimeStamp {
      	return MultiWorkerID(0), fmt.Errorf("Clock moved backwards. Refusing to generate id for %d milliseconds", n.lastTimeStamp - now)
    }

	if n.lastTimeStamp == now {
		n.sequence = (n.sequence + 1) & sequenceMask

		if n.sequence == 0 {
			now = snowflake.TilNextMillis(n.lastTimeStamp)
		}
	} else {
		n.sequence = 0
	}

	n.lastTimeStamp = now

	id = MultiWorkerID((now-snowflake.TW_EPOCH)<<timestampLeftShift | 
		(n.datacenterId << datacenterIdShift) |
		(n.nodeID << nodeIdShift) |
		(n.sequence))

	n.mu.Unlock()
	return
}

func (f MultiWorkerID) Int64() int64 {
	return int64(f)
}

func (f MultiWorkerID) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// Bytes returns a byte slice of the snowflake ID
func (f MultiWorkerID) Bytes() []byte {
	return []byte(f.String())
}

// Time returns an int64 unix timestamp of the snowflake ID time
func (f MultiWorkerID) Time() int64 {
	return (int64(f) >> timestampLeftShift) + snowflake.TW_EPOCH
}

// Node returns an int64 of the snowflake ID node number
func (f MultiWorkerID) Node() int64 {
	return int64(f) & nodeMask >> nodeIdShift
}

// Node returns an int64 of the snowflake ID node number
func (f MultiWorkerID) DataCenter() int64 {
	return int64(f) & datacenterMask >> datacenterIdShift
}

// Step returns an int64 of the snowflake step (or sequence) number
func (f MultiWorkerID) Sequence() int64 {
	return int64(f) & sequenceMask
}

// MarshalJSON returns a json byte array string of the snowflake ID.
func (f MultiWorkerID) MarshalJSON() ([]byte, error) {
	buff := make([]byte, 0, 22)
	buff = append(buff, '"')
	buff = strconv.AppendInt(buff, int64(f), 10)
	buff = append(buff, '"')
	return buff, nil
}

// UnmarshalJSON converts a json byte array of a snowflake ID into an ID type.
func (f *MultiWorkerID) UnmarshalJSON(b []byte) error {
	if len(b) < 3 || b[0] != '"' || b[len(b)-1] != '"' {
		return JSONSyntaxError{b}
	}

	i, err := strconv.ParseInt(string(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return err
	}

	*f = MultiWorkerID(i)
	return nil
}

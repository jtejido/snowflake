// Copyright 2010-2012 Twitter, Inc.
// license that can be found in the LICENSE file.

// package single implements the solution they outlined in their readme file:
// time - 41 bits (millisecond precision w/ a custom epoch gives us 69 years)
// configured machine id - 10 bits - gives us up to 1024 machines
// sequence number - 12 bits - rolls over every 4096 per machine (with protection to avoid rollover in the same ms)
package single

import (
	"github.com/jtejido/snowflake"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	nodeIdBits uint8 = 10
	maxNodeId   int64 = -1 ^ (-1 << nodeIdBits)

	sequenceBits uint8 = 12
	maxSequenceBits uint8 = -1 ^ (-1 << sequenceBits)

	nodeIdShift uint8 = sequenceBits
	nodeMask  int64 = maxNodeId << sequenceBits

	timestampLeftShift uint8 = sequenceBits + nodeIdBits

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
	ErrForwardTime = errors.New("You wish to generate new ID. use NextID() instead.")
)
type singleWorker struct {
	mu   sync.RWMutex
	lastTimeStamp int64
	nodeID int64
	sequence int64
}

type SingleWorkerID int64

func New(nodeID int64) (*singleWorker, error) {

	if (nodeID > maxNodeId || nodeID < 0) {
		return nil, ErrInvalidNode
	}

	return &singleWorker{
		lastTimeStamp: 0,
		nodeID: nodeID,
		sequence: 0,
	}, nil
}

// GetID gets the corresponding ID given a timestamp and sequence
func (n *singleWorker) GetID(t time.Time, s int64) (id SingleWorkerID, err error) {
 	
 	n.mu.Lock()

	now := t.UnixNano() / 1000000

	if n.lastTimeStamp != 0 && now > n.lastTimeStamp {
      	return SingleWorkerID(0), ErrForwardTime
    }

    // no prior knowledge of how many times an id has been generated for the same ms, store id.sequence() to be used here
    // TO-DO: get all IDs up to maxSequenceBits? (O(n) where n is maxSequenceBits)
	id = SingleWorkerID((now-snowflake.TW_EPOCH)<<timestampLeftShift | 
		(n.nodeID << nodeIdShift) |
		s)


	n.mu.Unlock()

	return
}

// LastTimeStamp returns last time.Time
func (n singleWorker) LastTimeStamp() time.Time {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return time.Unix(0, n.lastTimeStamp * 1000000)
}

// NextID creates and returns a unique snowflake ID
func (n *singleWorker) NextID() (id SingleWorkerID, err error) {

	n.mu.Lock()

	now := time.Now().UnixNano() / 1000000

	if now < n.lastTimeStamp {
      	return SingleWorkerID(0), fmt.Errorf("Clock moved backwards. Refusing to generate id for %d milliseconds", n.lastTimeStamp - now)
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

	id = SingleWorkerID((now-snowflake.TW_EPOCH)<<timestampLeftShift | 
		(n.nodeID << nodeIdShift) |
		n.sequence)

	n.mu.Unlock()
	return
}

func (f SingleWorkerID) Int64() int64 {
	return int64(f)
}

func (f SingleWorkerID) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// Bytes returns a byte slice of the snowflake ID
func (f SingleWorkerID) Bytes() []byte {
	return []byte(f.String())
}

// Time returns an int64 unix timestamp of the snowflake ID time
func (f SingleWorkerID) Time() int64 {
	return (int64(f) >> timestampLeftShift) + snowflake.TW_EPOCH
}

// Node returns an int64 of the snowflake ID node number
func (f SingleWorkerID) Node() int64 {
	return int64(f) & nodeMask >> nodeIdShift
}

// Step returns an int64 of the snowflake sequence number
func (f SingleWorkerID) Sequence() int64 {
	return int64(f) & sequenceMask
}

// not applicable here
func (f SingleWorkerID) DataCenter() int64 {
	return int64(0)
}

func (f SingleWorkerID) MarshalJSON() ([]byte, error) {
	buff := make([]byte, 0, 22)
	buff = append(buff, '"')
	buff = strconv.AppendInt(buff, int64(f), 10)
	buff = append(buff, '"')
	return buff, nil
}

func (f *SingleWorkerID) UnmarshalJSON(b []byte) error {
	if len(b) < 3 || b[0] != '"' || b[len(b)-1] != '"' {
		return JSONSyntaxError{b}
	}

	i, err := strconv.ParseInt(string(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return err
	}

	*f = SingleWorkerID(i)
	return nil
}

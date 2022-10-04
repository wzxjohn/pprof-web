package main

import "sync/atomic"

var id atomic.Int32

func nextId() int {
	_id := id.Add(1)
	if _id <= 65000 {
		return int(_id)
	}
	swapped := id.CompareAndSwap(_id, 1)
	if swapped {
		return 1
	}
	return nextId()
}

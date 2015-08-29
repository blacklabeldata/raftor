package main

import "errors"

var (
	ErrCanceled               = errors.New("raft: request cancelled")
	ErrTimeout                = errors.New("raft: request timed out")
	ErrTimeoutDueToLeaderFail = errors.New("raft: request timed out, possibly due to previous leader failure")
)

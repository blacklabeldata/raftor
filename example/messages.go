package main

type UserMessageType byte

const (
	UnknownMessageType UserMessageType = iota
	RaftMessageType
	TimestampMessageType
	AckMessageType
)

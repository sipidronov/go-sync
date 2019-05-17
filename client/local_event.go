package main

import ()

type ChangeType uint8

const (
	Modified ChangeType = iota
	Deleted
	Unsupported
	// TODO: Attr & extended attrs change?
)

type Event struct {
	Type ChangeType
	Path string
}

package main

import "time"

type Request struct {
	ID      string
	Method  string
	Path    string
	Headers map[string][]string
	Body    string

	CreatedAt time.Time
}

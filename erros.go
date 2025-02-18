package main

import "errors"

var (
	ErrMissingSeries    = errors.New("missing series")
	ErrMissingTitle     = errors.New("missing tittle")
	ErrMissingComicInfo = errors.New("missing comicinfo.xml")
)

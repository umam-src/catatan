package ui

import "embed"

// Assets adalah antarmuka web yang ditanam ke dalam program.
//go:embed dist/*
var Assets embed.FS

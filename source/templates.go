package main

import "embed"

//go:embed templ/*
var templateFS embed.FS

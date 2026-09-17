package webui

import "embed"

//go:embed all:dist
var Dist embed.FS

//go:embed all:shop
var Shop embed.FS

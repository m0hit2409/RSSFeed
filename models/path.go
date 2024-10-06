package models

type PathType int

const (
	TypeFile PathType = 1
	TypeUrl  PathType = 0
)

type Path struct {
	PathType PathType
	Path     string
}

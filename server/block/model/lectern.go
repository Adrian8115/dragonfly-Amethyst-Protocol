package model

import (
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/block/cube"
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/world"
)

// Lectern is a model used by lecterns.
type Lectern struct{}

// BBox ...
func (Lectern) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.9, 1)}
}

// FaceSolid ...
func (Lectern) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

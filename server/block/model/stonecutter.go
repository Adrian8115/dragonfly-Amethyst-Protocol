package model

import (
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/block/cube"
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/world"
)

// Stonecutter is a model used by stonecutters.
type Stonecutter struct{}

// BBox ...
func (Stonecutter) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.5625, 1)}
}

// FaceSolid ...
func (Stonecutter) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

package block

import (
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/block/cube"
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/item"
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/world"
)

// Placer represents an entity that is able to place a block at a specific position in the world.
type Placer interface {
	item.User
	PlaceBlock(pos cube.Pos, b world.Block, ctx *item.UseContext)
}

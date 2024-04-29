package session

import (
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/block/cube"
	"github.com/Adrian8115/gophertunnel-Amethyst-Protocol/minecraft/protocol/packet"
)

// BlockPickRequestHandler handles the BlockPickRequest packet.
type BlockPickRequestHandler struct{}

// Handle ...
func (b BlockPickRequestHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.BlockPickRequest)
	s.c.PickBlock(cube.Pos{int(pk.Position.X()), int(pk.Position.Y()), int(pk.Position.Z())})
	return nil
}

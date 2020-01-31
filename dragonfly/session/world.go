package session

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/action"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/state"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/chunk"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/particle"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync/atomic"
)

// handleRequestChunkRadius ...
func (s *Session) handleRequestChunkRadius(pk *packet.RequestChunkRadius) error {
	if pk.ChunkRadius > s.maxChunkRadius {
		pk.ChunkRadius = s.maxChunkRadius
	}
	atomic.StoreInt32(&s.chunkRadius, pk.ChunkRadius)

	s.chunkLoader.Load().(*world.Loader).ChangeRadius(int(pk.ChunkRadius))

	s.writePacket(&packet.ChunkRadiusUpdated{ChunkRadius: s.chunkRadius})
	return nil
}

// SendNetherDimension sends the player to the nether dimension
func (s *Session) SendNetherDimension() {
	s.writePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionNether,
		Position:  mgl32.Vec3{},
		Respawn:   false,
	})
}

// SendEndDimension sends the player to the end dimension
func (s *Session) SendEndDimension() {
	s.writePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionEnd,
		Position:  mgl32.Vec3{},
		Respawn:   false,
	})
}

// SendNetherDimension sends the player to the overworld dimension
func (s *Session) SendOverworldDimension() {
	s.writePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionOverworld,
		Position:  mgl32.Vec3{},
		Respawn:   false,
	})
}

// ViewChunk ...
func (s *Session) ViewChunk(pos world.ChunkPos, c *chunk.Chunk) {
	data := chunk.NetworkEncode(c)

	count := 16
	for y := 15; y >= 0; y-- {
		if data.SubChunks[y] == nil {
			count--
			continue
		}
		break
	}
	for y := 0; y < count; y++ {
		if data.SubChunks[y] == nil {
			_ = s.chunkBuf.WriteByte(chunk.SubChunkVersion)
			// We write zero here, meaning the sub chunk has no block storages: The sub chunk is completely
			// empty.
			_ = s.chunkBuf.WriteByte(0)
			continue
		}
		_, _ = s.chunkBuf.Write(data.SubChunks[y])
	}
	_, _ = s.chunkBuf.Write(data.Data2D)
	_, _ = s.chunkBuf.Write(data.BlockNBT)

	s.writePacket(&packet.LevelChunk{
		ChunkX:        pos[0],
		ChunkZ:        pos[1],
		SubChunkCount: uint32(count),
		RawPayload:    append([]byte(nil), s.chunkBuf.Bytes()...),
	})
	s.chunkBuf.Reset()
}

// ViewEntity ...
func (s *Session) ViewEntity(e world.Entity) {
	if s.entityRuntimeID(e) == selfEntityRuntimeID {
		return
	}
	var runtimeID uint64

	s.entityMutex.Lock()
	_, controllable := e.(Controllable)

	if id, ok := s.entityRuntimeIDs[e]; ok && controllable {
		runtimeID = id
	} else {
		runtimeID = atomic.AddUint64(&s.currentEntityRuntimeID, 1)
		s.entityRuntimeIDs[e] = runtimeID
		s.entities[runtimeID] = e
	}
	s.entityMutex.Unlock()

	switch v := e.(type) {
	case Controllable:
		s.writePacket(&packet.AddPlayer{
			UUID:            v.UUID(),
			Username:        v.Name(),
			EntityUniqueID:  int64(runtimeID),
			EntityRuntimeID: runtimeID,
			Position:        e.Position(),
			Pitch:           e.Pitch(),
			Yaw:             e.Yaw(),
			HeadYaw:         e.Yaw(),
		})
	default:
		s.writePacket(&packet.AddActor{
			EntityUniqueID:  int64(runtimeID),
			EntityRuntimeID: runtimeID,
			// TODO: Add methods for entity types.
			EntityType: "",
			Position:   e.Position(),
			Pitch:      e.Pitch(),
			Yaw:        e.Yaw(),
			HeadYaw:    e.Yaw(),
		})
	}
}

// HideEntity ...
func (s *Session) HideEntity(e world.Entity) {
	if s.entityRuntimeID(e) == selfEntityRuntimeID {
		return
	}

	s.entityMutex.Lock()
	id, ok := s.entityRuntimeIDs[e]
	if _, controllable := e.(Controllable); !controllable {
		delete(s.entityRuntimeIDs, e)
		delete(s.entities, s.entityRuntimeIDs[e])
	}
	s.entityMutex.Unlock()
	if !ok {
		// The entity was already removed some other way. We don't need to send a packet.
		return
	}
	s.writePacket(&packet.RemoveActor{EntityUniqueID: int64(id)})
}

// ViewEntityMovement ...
func (s *Session) ViewEntityMovement(e world.Entity, deltaPos mgl32.Vec3, deltaYaw, deltaPitch float32) {
	id := s.entityRuntimeID(e)

	if id == selfEntityRuntimeID {
		return
	}

	switch e.(type) {
	case Controllable:
		s.writePacket(&packet.MovePlayer{
			EntityRuntimeID: id,
			Position:        e.Position().Add(deltaPos),
			Pitch:           e.Pitch() + deltaPitch,
			Yaw:             e.Yaw() + deltaYaw,
			HeadYaw:         e.Yaw() + deltaYaw,
		})
	default:
		s.writePacket(&packet.MoveActorAbsolute{
			EntityRuntimeID: id,
			Position:        e.Position().Add(deltaPos),
			Rotation:        mgl32.Vec3{e.Pitch() + deltaPitch, e.Yaw() + deltaYaw},
		})
	}
}

// ViewTime ...
func (s *Session) ViewTime(time int) {
	s.writePacket(&packet.SetTime{Time: int32(time)})
}

// ViewEntityTeleport ...
func (s *Session) ViewEntityTeleport(e world.Entity, position mgl32.Vec3) {
	id := s.entityRuntimeID(e)

	if id == selfEntityRuntimeID {
		s.chunkLoader.Load().(*world.Loader).Move(position)
	}

	switch e.(type) {
	case Controllable:
		s.writePacket(&packet.MovePlayer{
			EntityRuntimeID: id,
			Position:        position,
			Pitch:           e.Pitch(),
			Yaw:             e.Yaw(),
			HeadYaw:         e.Yaw(),
			Mode:            packet.MoveModeTeleport,
		})
	default:
		s.writePacket(&packet.MoveActorAbsolute{
			EntityRuntimeID: id,
			Position:        position,
			Rotation:        mgl32.Vec3{e.Pitch(), e.Yaw()},
			Flags:           packet.MoveFlagTeleport,
		})
	}
}

// ViewEntityItems ...
func (s *Session) ViewEntityItems(e world.Entity) {
	c, ok := e.(item.Carrier)
	if !ok {
		return
	}

	if s.entityRuntimeID(e) == selfEntityRuntimeID {
		// Don't view the items of the entity if the entity is the Controllable of the session.
		return
	}
	mainHand, offHand := c.HeldItems()
	runtimeID := s.entityRuntimeID(e)

	// Show the main hand item.
	s.writePacket(&packet.MobEquipment{
		EntityRuntimeID: runtimeID,
		NewItem:         stackFromItem(mainHand),
	})
	// Show the off-hand item.
	s.writePacket(&packet.MobEquipment{
		EntityRuntimeID: runtimeID,
		NewItem:         stackFromItem(offHand),
		WindowID:        protocol.WindowIDOffHand,
	})
}

// ViewParticle ...
func (s *Session) ViewParticle(pos mgl32.Vec3, p particle.Particle) {
	switch pa := p.(type) {
	case particle.BlockBreak:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.EventParticleDestroy,
			Position:  pos,
			EventData: int32(s.blockRuntimeID(pa.Block)),
		})
	}
}

// ViewSound ...
func (s *Session) ViewSound(pos mgl32.Vec3, soundType sound.Sound) {
	switch so := soundType.(type) {
	case sound.BlockPlace:
		s.writePacket(&packet.LevelSoundEvent{
			SoundType:  packet.SoundEventPlace,
			Position:   pos,
			ExtraData:  int32(s.blockRuntimeID(so.Block)),
			EntityType: ":",
		})
	}
}

// ViewBlockUpdate ...
func (s *Session) ViewBlockUpdate(pos block.Position, b block.Block) {
	runtimeID, _ := block.RuntimeID(b)
	s.writePacket(&packet.UpdateBlock{
		Position:          protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])},
		NewBlockRuntimeID: runtimeID,
		Flags:             packet.BlockUpdateNetwork,
	})
}

// ViewEntityAction ...
func (s *Session) ViewEntityAction(e world.Entity, a action.Action) {
	switch a.(type) {
	case action.SwingArm:
		if _, ok := e.(Controllable); ok {
			s.writePacket(&packet.Animate{
				ActionType:      packet.AnimateActionSwingArm,
				EntityRuntimeID: s.entityRuntimeID(e),
			})
			return
		}
		s.writePacket(&packet.ActorEvent{
			EntityRuntimeID: s.entityRuntimeID(e),
			EventType:       packet.ActorEventArmSwing,
		})
	}
}

// ViewEntityState ...
func (s *Session) ViewEntityState(e world.Entity, states []state.State) {
	m := defaultEntityMetadata()
	for _, eState := range states {
		switch eState.(type) {
		case state.Sneaking:
			m.setFlag(dataKeyFlags, dataFlagSneaking)
		case state.Sprinting:
			m.setFlag(dataKeyFlags, dataFlagSprinting)
		case state.Breathing:
			m.setFlag(dataKeyFlags, dataFlagBreathing)
		}
	}
	s.writePacket(&packet.SetActorData{
		EntityRuntimeID: s.entityRuntimeID(e),
		EntityMetadata:  m,
	})
}

// blockRuntimeID returns the runtime ID of the block passed.
func (s *Session) blockRuntimeID(b block.Block) uint32 {
	id, _ := block.RuntimeID(b)
	return id
}

// entityRuntimeID returns the runtime ID of the entity passed.
func (s *Session) entityRuntimeID(e world.Entity) uint64 {
	s.entityMutex.RLock()
	id, _ := s.entityRuntimeIDs[e]
	s.entityMutex.RUnlock()
	return id
}

// entityFromRuntimeID attempts to return an entity by its runtime ID. False is returned if no entity with the
// ID could be found.
func (s *Session) entityFromRuntimeID(id uint64) (world.Entity, bool) {
	s.entityMutex.RLock()
	e, ok := s.entities[id]
	s.entityMutex.RUnlock()
	return e, ok
}
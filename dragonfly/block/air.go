package block

// Air is the block present in otherwise empty space.
type Air struct{}

func (Air) Name() string {
	return "Air"
}

func (Air) Minecraft() (name string, properties map[string]interface{}) {
	return "minecraft:air", nil
}
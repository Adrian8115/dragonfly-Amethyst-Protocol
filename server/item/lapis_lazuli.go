package item

import "github.com/Adrian8115/gophertunnel-Amethyst-Protocol/minecraft/text"

// LapisLazuli is a mineral used for enchanting and decoration.
type LapisLazuli struct{}

// EncodeItem ...
func (LapisLazuli) EncodeItem() (name string, meta int16) {
	return "minecraft:lapis_lazuli", 0
}

// TrimMaterial ...
func (LapisLazuli) TrimMaterial() string {
	return "lapis"
}

// MaterialColour ...
func (LapisLazuli) MaterialColour() string {
	return text.Lapis
}

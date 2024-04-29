package enchantment

import (
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/item"
	"github.com/Adrian8115/dragonfly-Amethyst-Protocol/server/world"
)

// AquaAffinity is a helmet enchantment that increases underwater mining speed.
type AquaAffinity struct{}

// Name ...
func (AquaAffinity) Name() string {
	return "Aqua Affinity"
}

// MaxLevel ...
func (AquaAffinity) MaxLevel() int {
	return 1
}

// Cost ...
func (AquaAffinity) Cost(int) (int, int) {
	return 1, 41
}

// Rarity ...
func (AquaAffinity) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithEnchantment ...
func (AquaAffinity) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (AquaAffinity) CompatibleWithItem(i world.Item) bool {
	h, ok := i.(item.HelmetType)
	return ok && h.Helmet()
}

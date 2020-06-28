package inventorytype

type InventoryType string

const (
	NonEquip InventoryType = "NON_EQUIP"
	Bag      InventoryType = "BAG"

	Ranged      InventoryType = "RANGED"
	RangedRight InventoryType = "RANGEDRIGHT"
	TwoWeapon   InventoryType = "TWOWEAPON"
	Weapon      InventoryType = "WEAPON"

	Head     InventoryType = "HEAD"
	Neck     InventoryType = "NECK"
	Shoulder InventoryType = "SHOULDER"
	Chest    InventoryType = "CHEST"
	Wrist    InventoryType = "WRIST"
	Hand     InventoryType = "HAND"
	Waist    InventoryType = "WAIST"
	Legs     InventoryType = "LEGS"
	Feet     InventoryType = "FEET"
	Finger   InventoryType = "FINGER"
	Trinket  InventoryType = "TRINKET"
)

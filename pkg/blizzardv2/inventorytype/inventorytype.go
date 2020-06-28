package inventorytype

type InventoryType string

const (
	NonEquip InventoryType = "NON_EQUIP"
	Bag      InventoryType = "BAG"
	Tabard   InventoryType = "TABARD"

	Ranged      InventoryType = "RANGED"
	RangedRight InventoryType = "RANGEDRIGHT"
	Ammo        InventoryType = "AMMO"
	TwoWeapon   InventoryType = "TWOWEAPON"
	Weapon      InventoryType = "WEAPON"
	Shield      InventoryType = "SHIELD"
	OffHand                   = "HOLDABLE"

	Head     InventoryType = "HEAD"
	Neck     InventoryType = "NECK"
	Shoulder InventoryType = "SHOULDER"
	Cloak    InventoryType = "CLOAK"
	Chest    InventoryType = "CHEST"
	Robe     InventoryType = "ROBE"
	Shirt    InventoryType = "BODY"
	Wrist    InventoryType = "WRIST"
	Hand     InventoryType = "HAND"
	Waist    InventoryType = "WAIST"
	Legs     InventoryType = "LEGS"
	Feet     InventoryType = "FEET"
	Finger   InventoryType = "FINGER"
	Trinket  InventoryType = "TRINKET"
)

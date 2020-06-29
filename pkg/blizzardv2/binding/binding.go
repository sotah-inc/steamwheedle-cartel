package binding

type Binding string

const (
	BindOnEquip   Binding = "ON_EQUIP"
	BindOnAccount Binding = "TO_BNETACCOUNT"
	BindOnPickup  Binding = "ON_ACQUIRE"
)

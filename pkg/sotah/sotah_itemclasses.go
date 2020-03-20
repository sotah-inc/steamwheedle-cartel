package sotah

type ItemClasses struct {
	Classes []ItemClass `json:"classes"`
}

type ItemClassClass int

type ItemClass struct {
	Class      ItemClassClass `json:"class"`
	Name       string         `json:"name"`
	SubClasses []SubItemClass `json:"subclasses"`
}

type ItemSubClassClass int

type SubItemClass struct {
	SubClass ItemSubClassClass `json:"subclass"`
	Name     string            `json:"name"`
}

package blizzardv2

type PetIds []PetId

func (ids PetIds) ToUniqueMap() PetIdsMap {
	out := PetIdsMap{}
	for _, id := range ids {
		out[id] = struct{}{}
	}

	return out
}

type PetIdsMap map[PetId]struct{}

func (idsMap PetIdsMap) Exists(id PetId) bool {
	_, ok := idsMap[id]

	return ok
}

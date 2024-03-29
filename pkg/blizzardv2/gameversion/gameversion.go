package gameversion

import "errors"

type GameVersion string

const (
	Classic GameVersion = "classic"
	Retail  GameVersion = "retail"
)

type List []GameVersion

func (l List) Includes(providedVersion GameVersion) bool {
	for _, version := range l {
		if version == providedVersion {
			return true
		}
	}

	return false
}

type VersionNamespaceMap map[GameVersion]string

var DynamicNamespaceMap = VersionNamespaceMap{
	Classic: "dynamic-classic-us",
	Retail:  "dynamic-us",
}

var StaticNamespaceMap = VersionNamespaceMap{
	Classic: "static-classic-us",
	Retail:  "static-us",
}

func (vnMap VersionNamespaceMap) Resolve(version GameVersion) (string, error) {
	found, ok := vnMap[version]
	if !ok {
		return "", errors.New("could not resolve namespace")
	}

	return found, nil
}

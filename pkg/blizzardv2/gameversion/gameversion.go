package gameversion

import "errors"

type GameVersion string

const (
	Classic GameVersion = "classic"
	Retail  GameVersion = "retail"
)

type List []GameVersion

var GameVersions = List{Classic, Retail}

type VersionNamespaceMap map[GameVersion]string

var DynamicNamespaceMap = VersionNamespaceMap{
	Classic: "dynamic-classic-us",
	Retail:  "dynamic-us",
}

func (vnMap VersionNamespaceMap) Resolve(version GameVersion) (string, error) {
	found, ok := vnMap[version]
	if !ok {
		return "", errors.New("could not resolve namespace")
	}

	return found, nil
}

// !Trustno1Cox

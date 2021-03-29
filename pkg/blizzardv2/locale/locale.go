package locale

import (
	"encoding/json"
	"errors"
)

type Locale string

func (locale Locale) IsZero() bool {
	return len(locale) == 0
}

func NewMapping(data []byte) (Mapping, error) {
	out := Mapping{}
	if err := json.Unmarshal(data, &out); err != nil {
		return Mapping{}, err
	}

	return out, nil
}

type Mapping map[Locale]string

func (m Mapping) ResolveDefaultName() string {
	return m.FindOr(EnUS, "")
}

func (m Mapping) IsZero() bool {
	return len(m) == 0
}

func (m Mapping) EncodeForStorage() ([]byte, error) {
	return json.Marshal(m)
}

func (m Mapping) Find(locale Locale) (string, error) {
	found, ok := m[locale]
	if !ok {
		return "", errors.New("could not resolve locale")
	}

	return found, nil
}

func (m Mapping) FindOr(locale Locale, defaultValue string) string {
	found, ok := m[locale]
	if !ok {
		return defaultValue
	}

	return found
}

const (
	EnUS Locale = "en_US"
	EsMX Locale = "es_MX"
	PtBR Locale = "pt_BR"
	DeDE Locale = "de_DE"
	EnGB Locale = "en_GB"
	EsES Locale = "es_ES"
	FrFR Locale = "fr_FR"
	ItIT Locale = "it_IT"
	RuRU Locale = "ru_RU"
	KoKR Locale = "ko_KR"
	ZhTW Locale = "zh_TW"
	ZhCN Locale = "zh_CN"
)

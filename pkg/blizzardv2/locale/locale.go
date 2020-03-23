package locale

type Locale string
type Mapping map[Locale]string

func (m Mapping) ResolveDefaultName() string {
	found, ok := m[EnUS]
	if !ok {
		return "NO NAME FOUND"
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

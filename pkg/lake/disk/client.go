package disk

func NewClient(cacheDir string) Client { return Client{CacheDir: cacheDir} }

type Client struct {
	CacheDir string
}

package disk

func NewClient(cacheDir string) Client { return Client{cacheDir: cacheDir} }

type Client struct {
	cacheDir string
}

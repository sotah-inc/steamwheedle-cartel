package disk

import "encoding/json"

func (client Client) GetEncodedItemClasses() ([]byte, error) {
	itemClasses, err := client.resolveItemClasses()
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal(itemClasses)
}

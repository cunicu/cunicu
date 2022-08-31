package util

import "encoding/json"

func ReIndentJSON(j []byte, prefix, indent string) ([]byte, error) {
	u := map[string]any{}

	if err := json.Unmarshal(j, &u); err != nil {
		return nil, err
	}

	return json.MarshalIndent(u, prefix, indent)
}

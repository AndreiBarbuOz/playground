package main

import (
	"encoding/json"
	"fmt"
)

// unmarshallUnstructured unmarshalls the body received as an argument and attempts to extract the list of tags from the content
// return error if the body doesn't match expected schema
func unmarshallUnstructured(body []byte) ([]string, error) {
	var result map[string]interface{}

	var tags []string

	err := json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshall body: %w", err)
	}

	for key, value := range result {
		// ondemand images are not part of the required list
		if key == "ondemand" {
			continue
		}

		if list, ok := value.([]interface{}); ok {
			for _, inner := range list {
				if tag, ok := inner.(string); ok {
					tags = append(tags, tag)
				}
			}
		} else {
			return nil, fmt.Errorf("expected type list: \"%v\"", list)
		}
	}

	return tags, nil
}

package mysql

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(v any) error {
	jsonText, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonText))

	return nil
}

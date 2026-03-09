package stage

import "fmt"

func validateImageLists(originals, transfers []string) error {
	if len(originals) != len(transfers) {
		return fmt.Errorf("image list mismatch: %d originals, %d transfers", len(originals), len(transfers))
	}

	return nil
}

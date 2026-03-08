package stage

func progressMessage(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}

	return plural
}

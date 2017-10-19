package dockerversion

func findLatestEdgeVersion(mirrorURL string) (Version, error) {
	results, err := ListVersions(mirrorURL, Edge)
	if err != nil {
		return Version{}, err
	}

	last := len(results) - 1
	return results[last], nil
}

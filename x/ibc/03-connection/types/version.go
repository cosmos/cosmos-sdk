package types

// GetCompatibleVersions returns an ordered set of compatible IBC versions for the
// caller chain's connection end.
func GetCompatibleVersions() []string {
	return []string{"1.0.0"}
}

// LatestVersion gets the latest version of a connection protocol
//
// CONTRACT: version array MUST be already sorted.
func LatestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}
	return versions[len(versions)-1]
}

// PickVersion picks the counterparty latest version that is matches the list
// of compatible versions for the connection.
func PickVersion(counterpartyVersions, compatibleVersions []string) string {

	n := len(counterpartyVersions)
	m := len(compatibleVersions)

	// aux hash maps to lookup already seen versions
	counterpartyVerLookup := make(map[string]bool)
	compatibleVerLookup := make(map[string]bool)

	// versions suported
	var supportedVersions []string

	switch {
	case n == 0 || m == 0:
		return ""
	case n == m:
		for i := n - 1; i >= 0; i-- {
			counterpartyVerLookup[counterpartyVersions[i]] = true
			compatibleVerLookup[compatibleVersions[i]] = true

			// check if we've seen any of the versions
			if _, ok := compatibleVerLookup[counterpartyVersions[i]]; ok {
				supportedVersions = append(supportedVersions, counterpartyVersions[i])
			}

			if _, ok := counterpartyVerLookup[compatibleVersions[i]]; ok {
				// TODO: check if the version is already in the array
				supportedVersions = append(supportedVersions, compatibleVersions[i])
			}
		}
	case n > m:
		for i := n - 1; i >= m; i-- {
			counterpartyVerLookup[compatibleVersions[i]] = true
		}

		for i := m - 1; i >= 0; i-- {
			counterpartyVerLookup[counterpartyVersions[i]] = true
			compatibleVerLookup[compatibleVersions[i]] = true

			// check if we've seen any of the versions
			if _, ok := compatibleVerLookup[counterpartyVersions[i]]; ok {
				supportedVersions = append(supportedVersions, counterpartyVersions[i])
			}

			if _, ok := counterpartyVerLookup[compatibleVersions[i]]; ok {
				supportedVersions = append(supportedVersions, compatibleVersions[i])
			}
		}

	case n < m:
		for i := m - 1; i >= n; i-- {
			compatibleVerLookup[compatibleVersions[i]] = true
		}

		for i := n - 1; i >= 0; i-- {
			counterpartyVerLookup[counterpartyVersions[i]] = true
			compatibleVerLookup[compatibleVersions[i]] = true

			// check if we've seen any of the versions
			if _, ok := compatibleVerLookup[counterpartyVersions[i]]; ok {
				supportedVersions = append(supportedVersions, counterpartyVersions[i])
			}

			if _, ok := counterpartyVerLookup[compatibleVersions[i]]; ok {
				supportedVersions = append(supportedVersions, compatibleVersions[i])
			}
		}
	}

	if len(supportedVersions) == 0 {
		return ""
	}

	// TODO: compare latest version before appending
	return supportedVersions[len(supportedVersions)-1]
}

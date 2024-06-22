package utils

func RemoveElements(original []string, toRemove []string) []string {
	var result []string
	for _, i := range original {
		remove := false
		for _, j := range toRemove {
			if i == j {
				remove = true
				break
			}
		}
		if !remove {
			result = append(result, i)
		}
	}
	return result
}

func RemoveDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if !encountered[elements[v]] {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

func StringInList(target string, list []string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

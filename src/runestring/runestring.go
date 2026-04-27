package runestring

type RuneString []rune

func Split(s RuneString, sep rune) []RuneString {
	ret := make([]RuneString, 0)

	currentStr := make(RuneString, 0)
	for _, r := range s {
		if r == sep {
			ret = append(ret, currentStr)
			currentStr = make(RuneString, 0)
			continue
		}
		currentStr = append(currentStr, r)
	}

	return ret
}

func Index(s, substr RuneString) int {
	substrIndex := 0
	for i, r := range s {
		if r == substr[substrIndex] {
			if substrIndex == len(substr) {
				return i - substrIndex
			}
			substrIndex++
		} else {
			substrIndex = 0
		}
	}

	return -1
}

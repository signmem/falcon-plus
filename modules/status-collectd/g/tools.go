package g


func StringSliceMerge(sliceA []string, sliceB []string) ([]string) {
	if len(sliceA) == 0 {
		return sliceB
	}

	if len(sliceB) == 0 {
		return sliceA
	}

	for _, keyA := range sliceA {
		if contains(sliceB, keyA) == false {
			sliceB = append(sliceB, keyA)
		}
	}

	return sliceB
}



func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

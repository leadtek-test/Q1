package command

func cloneStrings(input []string) []string {
	if input == nil {
		return nil
	}
	output := make([]string, len(input))
	copy(output, input)
	return output
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}

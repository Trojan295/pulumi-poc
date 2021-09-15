package utils

func StringArrayOutputFunc(args []interface{}) []string {
	result := make([]string, 0)
	for _, el := range args {
		s := el.(string)
		result = append(result, s)
	}
	return result
}

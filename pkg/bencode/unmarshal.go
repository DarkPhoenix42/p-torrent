package bencode

func UnMarshal(data []byte) (any, error) {
	switch data[0] {
	case 'i':
		return unmarshalInt(data)

	case 'l':
		return unmarshalList(data)

	case 'd':
		return unmarshalDict(data)

	default:
		return unmarshalString(data)
	}
}

func unmarshalInt(data []byte) (int, error) {
	return 0, nil
}

func unmarshalList(data []byte) ([]any, error) {
	return nil, nil
}

func unmarshalDict(data []byte) (map[string]any, error) {
	return nil, nil
}

func unmarshalString(data []byte) (string, error) {
	return "", nil
}

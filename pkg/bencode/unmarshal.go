package bencode

import (
	"bytes"
	"fmt"
	"strconv"
)

func UnMarshal(data *[]byte) (any, error) {
	switch (*data)[0] {
	case 'i':
		return unmarshalInt(data)

	case 'l':
		return unmarshalList(data)

	case 'd':
		return unmarshalDict(data)

	case 'e':
		return nil, nil

	default:
		return unmarshalString(data)
	}
}

func unmarshalInt(data *[]byte) (any, error) {
	end_idx := bytes.IndexByte(*data, 'e')
	if end_idx == -1 {
		return nil, fmt.Errorf("invalid bencode data")
	}

	int_data := (*data)[1:end_idx]
	int_val, err := strconv.Atoi(string(int_data))
	if err != nil {
		return nil, fmt.Errorf("invalid value for int")
	}

	*data = (*data)[end_idx+1:]
	return int_val, nil
}

func unmarshalList(data *[]byte) ([]any, error) {
	list := []any{}

	*data = (*data)[1:]
	for (*data)[0] != 'e' {
		val, err := UnMarshal(data)
		if err != nil {
			return nil, err
		}

		list = append(list, val)
	}

	*data = (*data)[1:]
	return list, nil
}

func unmarshalDict(data *[]byte) (map[string]any, error) {
	dict := map[string]any{}

	*data = (*data)[1:]
	for (*data)[0] != 'e' {
		key, err := unmarshalString(data)
		key_str := string(key)
		if err != nil {
			return nil, err
		}

		val, err := UnMarshal(data)
		if err != nil {
			return nil, err
		}

		dict[key_str] = val
	}

	*data = (*data)[1:]
	return dict, nil
}

func unmarshalString(data *[]byte) (string, error) {
	colon_idx := bytes.IndexByte(*data, ':')
	if colon_idx == -1 {
		return "", fmt.Errorf("invalid bencode data")
	}

	str_len_data := (*data)[:colon_idx]
	str_len, err := strconv.Atoi(string(str_len_data))
	if err != nil {
		return "", fmt.Errorf("invalid value for string length")
	}

	str_data := (*data)[colon_idx+1 : colon_idx+1+str_len]
	*data = (*data)[colon_idx+1+str_len:]
	return string(str_data), nil
}

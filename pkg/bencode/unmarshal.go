package bencode

import (
	"bytes"
	"fmt"
	"strconv"
)

// TODO: Add support for arbitrary structs

func UnMarshal(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("invalid bencode data")
	}

	val, _, err := unmarshal(data, 0)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func unmarshal(data []byte, offset int) (any, int, error) {
	switch data[offset] {
	case 'i':
		return unmarshalInt(data, offset)
	case 'l':
		return unmarshalList(data, offset)

	case 'd':
		return unmarshalDict(data, offset)

	case 'e':
		return nil, 0, nil

	default:
		return unmarshalString(data, offset)
	}
}

func unmarshalInt(data []byte, offset int) (any, int, error) {
	end_idx := bytes.IndexByte(data[offset:], 'e')
	if end_idx == -1 {
		return nil, 0, fmt.Errorf("invalid bencode data")
	}

	int_data := data[offset+1 : offset+end_idx]
	int_val, err := strconv.Atoi(string(int_data))
	if err != nil {
		return nil, 0, fmt.Errorf("invalid value for int")
	}

	offset += end_idx + 1
	return int_val, offset, nil
}

func unmarshalList(data []byte, offset int) ([]any, int, error) {
	list := []any{}
	offset += 1

	for data[offset] != 'e' {
		val, new_offset, err := unmarshal(data, offset)
		if err != nil {
			return nil, 0, err
		}

		offset = new_offset
		list = append(list, val)
	}
	offset += 1
	return list, offset, nil
}

func unmarshalDict(data []byte, offset int) (map[string]any, int, error) {
	dict := map[string]any{}
	offset += 1

	for data[offset] != 'e' {
		key, new_offset, err := unmarshalString(data, offset)
		if err != nil {
			return nil, 0, err
		}

		key_str := string(key)
		offset = new_offset

		val, new_offset, err := unmarshal(data, offset)
		if err != nil {
			return nil, 0, err
		}

		offset = new_offset
		dict[key_str] = val
	}

	offset += 1
	return dict, offset, nil
}

func unmarshalString(data []byte, offset int) (string, int, error) {
	colon_idx := bytes.IndexByte(data[offset:], ':')
	if colon_idx == -1 {
		return "", 0, fmt.Errorf("invalid bencode data")
	}

	str_len_data := data[offset : offset+colon_idx]
	str_len, err := strconv.Atoi(string(str_len_data))
	if err != nil {
		return "", 0, fmt.Errorf("invalid value for string length")
	}
	str_data := data[offset+colon_idx+1 : offset+colon_idx+str_len+1]

	offset += colon_idx + str_len + 1
	return string(str_data), offset, nil
}

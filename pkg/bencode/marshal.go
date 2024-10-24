package bencode

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

// TODO: Add support for arbitrary structs

func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	switch value := v.(type) {

	case int:
		marshalInt(value, &buf)

	case string:
		marshalString(value, &buf)

	case []byte:
		marshalString(string(value), &buf)

	case []any:
		marshalList(value, &buf)

	case map[string]any:
		marshalDict(value, &buf)

	default:
		return nil, fmt.Errorf("unsupported type:  %T", value)
	}
	return buf.Bytes(), nil
}

func marshalInt(v int, buf *bytes.Buffer) {
	buf.WriteRune('i')
	buf.WriteString(strconv.Itoa(v))
	buf.WriteRune('e')
}

func marshalString(v string, buf *bytes.Buffer) {
	buf.WriteString(strconv.Itoa(len(v)))
	buf.WriteRune(':')
	buf.WriteString(v)
}

func marshalList(v []any, buf *bytes.Buffer) {
	buf.WriteRune('l')

	for _, e := range v {
		b, err := Marshal(e)
		if err != nil {
			panic(err)
		}
		buf.Write(b)
	}

	buf.WriteRune('e')
}

func marshalDict(v map[string]any, buf *bytes.Buffer) {
	buf.WriteRune('d')

	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		marshalString(k, buf)
		b, err := Marshal(v[k])
		if err != nil {
			panic(err)
		}
		buf.Write(b)
	}

	buf.WriteRune('e')
}

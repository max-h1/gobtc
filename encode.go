package main

import (
	"bufio"
	"fmt"
	"slices"
	"strconv"
)

type Encoder struct {
	buffer *bufio.Writer
}

func (encoder *Encoder) encodeInt(num int) error {
	out := "i" + strconv.Itoa(int(num)) + "e"

	_, err := encoder.buffer.WriteString(out)

	if err != nil {
		return err
	}

	return nil
}

func (encoder *Encoder) encodeString(str string) error {
	out := strconv.Itoa(len(str)) + ":" + string(str)

	_, err := encoder.buffer.WriteString(out)

	if err != nil {
		return err
	}

	return nil
}


func (encoder *Encoder) encodeList(list []any) error {
	encoder.buffer.WriteByte('l')

	for _, item := range list {
		err := encoder.encodeInterface(item)

		if err != nil { return err }
	}

	encoder.buffer.WriteByte('e')
	
	return nil
}

func (encoder *Encoder) encodeDict(dict map[string]any) error {
	encoder.buffer.WriteByte('d')

	keys := make([]string, 0, len(dict))

	for key := range dict {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	for _, key := range keys {
		err := encoder.encodeString(string(key))

		if err != nil { return err }

		err = encoder.encodeInterface(dict[key])

		if err != nil { return err }
	}

	err := encoder.buffer.WriteByte('e')

	if err != nil { return err }

	return nil
}


func (encoder *Encoder) encodeInterface(item any) error {
	switch item := item.(type) {
	case int:
		return encoder.encodeInt(item)
	case string:
		return encoder.encodeString(item)
	case []any:
		return encoder.encodeList(item)
	case map[string]any:
		return encoder.encodeDict(item)
	case *int:
		return encoder.encodeInt(*item)
	case *string:
		return encoder.encodeString(*item)
	case *[]any:
		return encoder.encodeList(*item)
	case *map[string]any:
		return encoder.encodeDict(*item)
	default:
		return fmt.Errorf("encoder - invalid type: %v", item)
	}
}

func (encoder *Encoder) Encode(dict map[string]any) error {

	err := encoder.encodeDict(dict)

	if err != nil {
		return err
	}

	encoder.buffer.Flush()

	return nil
}
package main

import (
	"bufio"
	"errors"
	"reflect"
	"strconv"
)

type Encoder struct {
	buffer *bufio.Writer
}

func (encoder *Encoder) encodeInt(num int) error {
	out := "i" + strconv.Itoa(num) + "e"

	_, err := encoder.buffer.WriteString(out)

	if err != nil {
		return err
	}

	return nil
}

func (encoder *Encoder) encodeString(str string) error {
	out := strconv.Itoa(len(str)) + ":" + str

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

	for key, val := range dict {
		err := encoder.encodeString(key)

		if err != nil { return err }

		err = encoder.encodeInterface(val)

		if err != nil { return err }
	}

	err := encoder.buffer.WriteByte('e')

	if err != nil { return err }

	return nil
}


func (encoder *Encoder) encodeInterface(item any) error {
	typ := reflect.TypeOf(item)
	kind := typ.Kind()

	switch kind {
	case reflect.Slice | reflect.Array:
		return encoder.encodeList(item.([]any))
	case reflect.Int:
		return encoder.encodeInt(item.(int))
	case reflect.String:
		return encoder.encodeString(item.(string))
	case reflect.Map:
		return encoder.encodeDict(item.(map[string]any))
	default:
		return errors.New("invalid type")
	}
}

func (encoder *Encoder) Encode(dict map[string]any) error {

	err := encoder.encodeDict(dict)

	if err != nil {
		return err
	}

	return nil
}
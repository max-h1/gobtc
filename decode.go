package main

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

type Decoder struct {
	buffer *bufio.Reader
}


func (decoder *Decoder) decodeString() (string, error) {
	rawlen, err := decoder.buffer.ReadBytes(':')
	
	if err != nil {
		return "", err
	}

	// Don't want delimiter
	length, err := strconv.Atoi(string(rawlen[:len(rawlen)-1]))

	if err != nil {
		return "", err
	}

	if length < 0 {
		return "", errors.New("negative length string")
	}

	str := make([]byte, length)

	_, err = io.ReadFull(decoder.buffer, str)

	if err != nil {
		return "", err
	}

	return string(str), nil
}

func (decoder *Decoder) decodeInt() (int, error) {
	raw, err := decoder.buffer.ReadBytes('e')
	
	if err != nil {
		return 0, err
	}

	num, err := strconv.Atoi(string(raw[:len(raw)-1]))

	if err != nil {
		return 0, err
	}
	
	return num, nil
}

func (decoder *Decoder) decodeList() ([]any, error) {
	list := make([]any, 0)
	for {
		next, err := decoder.buffer.ReadByte()

		if err != nil {
			return nil, err
		}

		if next == 'e' {
			break
		} else {
			decoder.buffer.UnreadByte()
		}

		item, err := decoder.decodeInterface()

		if err != nil {
			return nil, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (decoder *Decoder) decodeDict() (map[string]any, error) {
	dict := make(map[string]any)
	for {
		next, err := decoder.buffer.ReadByte()

		if err != nil {
			return nil, err
		}

		if next == 'e' {
			break
		} else {
			decoder.buffer.UnreadByte()
		}

		key, err := decoder.decodeString()

		if err != nil {
			return nil, err
		}

		val, err := decoder.decodeInterface()

		if err != nil {
			return nil, err
		}

		dict[key] = val
	}
	return dict, nil
}

func (decoder *Decoder) decodeInterface() (any, error) {
	next, err := decoder.buffer.ReadByte()
	
	if err != nil {
		return nil, err
	}

	switch {
	case next == 'i':
		return decoder.decodeInt()
	case next >= '0' && next <= '9':
		decoder.buffer.UnreadByte()
		return decoder.decodeString()
	case next == 'l':
		return decoder.decodeList()
	case next == 'd':
		return decoder.decodeDict()
	default:
		return nil, errors.New("invalid type")
	}
}

func (decoder *Decoder) Decode() (map[string]any, error) {

	firstByte, err := decoder.buffer.ReadByte()

	if err != nil {
		return nil, err
	}

	if firstByte != 'd' {
		return nil, errors.New("must start with a dictionary")
	}

	return decoder.decodeDict()
}
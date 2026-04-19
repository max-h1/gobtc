package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Decoder struct {
	data      []byte
	buffer    *bufio.Reader
	pos       int
	infoStart int
	infoEnd   int
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data:   data,
		buffer: bufio.NewReader(bytes.NewReader(data)),
		pos:    0,
	}
}

// func (decoder *Decoder) CalculateInfoHash() ([]byte, error) {
// 	var buf []byte

// 	_, err := io.ReadFull(decoder.buffer, buf)
// }

func (decoder *Decoder) decodeString() (string, error) {
	rawlen, err := decoder.buffer.ReadBytes(':')

	if err != nil {
		return "", err
	}
	decoder.pos += len(rawlen)

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
	decoder.pos += len(str)

	return string(str), nil
}

func (decoder *Decoder) decodeInt() (int, error) {
	raw, err := decoder.buffer.ReadBytes('e')

	if err != nil {
		return 0, err
	}
	decoder.pos += len(raw)

	num, err := strconv.Atoi(string(raw[:len(raw)-1]))

	if err != nil {
		return 0, err
	}

	return int(num), nil
}

func (decoder *Decoder) decodeList() ([]any, error) {
	list := make([]any, 0)
	for {
		next, err := decoder.buffer.ReadByte()

		if err != nil {
			return nil, err
		}
		decoder.pos++

		if next == 'e' {
			break
		} else {
			decoder.buffer.UnreadByte()
			decoder.pos--
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
		decoder.pos++

		if next == 'e' {
			break
		} else {
			decoder.buffer.UnreadByte()
			decoder.pos--
		}

		// keyStartPos := decoder.pos

		key, err := decoder.decodeString()

		if err != nil {
			return nil, err
		}

		if key == "info" && decoder.infoStart == 0 {
			// The next byte to be read is the start of the info dictionary ('d')
			decoder.infoStart = decoder.pos
		}

		val, err := decoder.decodeInterface()

		if err != nil {
			return nil, err
		}

		if key == "info" && decoder.infoEnd == 0 {
			decoder.infoEnd = decoder.pos
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
	decoder.pos++

	switch {
	case next == 'i':
		return decoder.decodeInt()
	case next >= '0' && next <= '9':
		decoder.buffer.UnreadByte()
		decoder.pos--
		return decoder.decodeString()
	case next == 'l':
		return decoder.decodeList()
	case next == 'd':
		return decoder.decodeDict()
	default:
		return nil, fmt.Errorf("decoder - invalid type: \"%c\"", next)
	}
}

func (decoder *Decoder) Decode() (map[string]any, error) {

	firstByte, err := decoder.buffer.ReadByte()

	if err != nil {
		return nil, err
	}
	decoder.pos++

	if firstByte != 'd' {
		return nil, errors.New("must start with a dictionary")
	}

	return decoder.decodeDict()
}

func (decoder *Decoder) InfoHash() ([]byte, error) {
	if decoder.infoStart == 0 || decoder.infoEnd == 0 || decoder.infoEnd <= decoder.infoStart {
		return nil, errors.New("info dictionary not found")
	}
	sum := sha1.Sum(decoder.data[decoder.infoStart:decoder.infoEnd])
	return sum[:], nil
}
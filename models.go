package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/structs"
)

type TrackerRequest struct {
	Info_hash 	[]byte `structs:"info_hash"`
	Peer_id 	[]byte `structs:"peer_id"`
	Port 		int `structs:"port"`
	Uploaded 	int `structs:"uploaded"`
	Downloaded 	int `structs:"downloaded"`
	Left 		int `structs:"left"`
}

type Metainfo struct {
	Announce 	string `structs:"announce"`
	Info 		Info `structs:"info_hash"`
}

type Info struct {
	Name 			string `structs:"name"`
	Piece_length 	int `structs:"piece_length"`
	Pieces 			string `structs:"pieces"`
	Length 			*int `structs:"length,omitempty"`
	Files 			[]File `structs:"files,omitempty"`
}

type File struct {
	Length 	int `structs:"length"`
	Path 	[]any `structs:"path"`
}

func BuildInfoHash(metainfo Metainfo) ([]byte, error) {
	var buf bytes.Buffer

	writer := bufio.NewWriter(&buf)

	encoder := Encoder{writer}

	info := metainfo.Info
	infoStruct := structs.Map(info)
	err := encoder.Encode(infoStruct)

	fmt.Println(buf)

	if err != nil { return nil, err }

	hasher := sha1.New()
	_, err = hasher.Write(buf.Bytes())

	if err != nil { return nil, err}

	hash := hasher.Sum(nil)

	hexHash := hex.EncodeToString(hash)

	fmt.Printf("info_hash: %v", hexHash)

	return hash, nil
}

func BuildTrackerRequest(infoHash []byte, port int, uploaded int, downloaded int, left int) (TrackerRequest, error) {
	var request TrackerRequest

	request.Info_hash = infoHash

	peer_id := make([]byte, 20)

	_, err := rand.Read(peer_id)

	if err != nil { return TrackerRequest{}, err }


	request.Peer_id = peer_id
	request.Port = port
	request.Uploaded = uploaded
	request.Downloaded = downloaded
	request.Left = left

	return request, nil
}

func escapeBytes(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '.' || c == '_' || c == '~' {
			sb.WriteByte(c)
		} else {
			sb.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return sb.String()
}

func BuildMetainfo(obj map[string]any) (Metainfo, error) {
    var meta Metainfo

    announce, ok := obj[string("announce")].(string)
    if !ok {
        return Metainfo{}, errors.New(`error building Metainfo struct: missing or invalid "announce"`)
    }
    meta.Announce = announce

    infoVal, ok := obj[string("info")].(map[string]any)
    if !ok {
        return Metainfo{}, errors.New(`error building Metainfo struct: missing or invalid "info"`)
    }
    info, err := BuildInfo(infoVal)
    if err != nil {
        return Metainfo{}, err
    }
    meta.Info = info

    return meta, nil
}

func BuildInfo(obj map[string]any) (Info, error) {
	var info Info

	if name, ok := obj[string("name")].(string); !ok {
		return Info{}, errors.New(`error building Info struct: missing or invalid "name"`)
	} else {
		info.Name = name
	}

	if piece_length, ok := obj[string("piece length")].(int); !ok {
		return Info{}, errors.New("error building Info struct: missing or invalid \"piece length\"")
	} else {
		info.Piece_length = piece_length
	}
	
	if pieces, ok := obj[string("pieces")].(string); !ok {
		return Info{}, errors.New("error building Info struct: missing or invalid \"pieces\"")
	} else {
		info.Pieces = pieces
	}
	
	length, lengthOk := obj[string("length")].(int)
	files, filesOk := obj[string("files")].([]any)

	switch {
	case lengthOk && filesOk:
		return Info{}, errors.New("error building Info struct: fields \"length\" and \"files\" cannot both be used")
	case !lengthOk && !filesOk:
		return Info{}, errors.New("error building Info struct: one of fields \"length\" and \"files\" are required")
	case lengthOk && !filesOk:
		info.Length = &length
	case !lengthOk && filesOk:
		builtFiles, err := BuildFiles(files)

		if err != nil { return Info{}, err}
		info.Files = builtFiles
	}
	return info, nil
}

func BuildFiles(obj []any) ([]File, error) {
	var files []File

	for _, item := range obj{
		var file File
		item, ok := item.(map[string]any)
		if !ok {
			return nil, errors.New("error building File struct: invalid list item")
		}
		if length, ok := item["length"].(int); !ok {
			return nil, errors.New("error building File struct: missing or invalid \"length\"")
		} else {
			file.Length = length
		}

		if path, ok := item["path"].([]any); !ok {
			return nil, errors.New("error building Files struct: missing or invalid \"path\"")
		} else {
			for _, v := range path {
				if _, ok := v.(string); !ok {
					return nil, errors.New("error buildings Files struct: item in field \"path\" is not a string")
				}
			}
			file.Path = path
		}

		files = append(files, file)
	}

	return files, nil
}
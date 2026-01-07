package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func PrintJSON(obj interface{}) {
	bytes, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(bytes))
}

func main() {
	infile, err := os.Open("test_data/ubuntu.torrent")

	if err != nil {
		panic(err)
	}

	defer infile.Close()

	reader := bufio.NewReader(infile)

	decoder := Decoder{reader}
	
	output, err := decoder.Decode()

	// PrintJSON(output)

	metainfo, err := BuildMetainfo(output)



	// fmt.Println(metainfo)

	req, err := BuildTrackerRequest(metainfo, 6881, 0, 0, 0)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(req)

	params := fmt.Sprintf(
		"info_hash=%s&peer_id=%s&port=6881&uploaded=0&downloaded=0&left=0&compact=1",
		escapeBytes(req.Info_hash),
		escapeBytes(req.Peer_id),
	)

	url := metainfo.Announce + "?" + params

	response, err := http.Get(url)

	body, _ := io.ReadAll(response.Body)
	fmt.Printf("%q\n", body)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)
}
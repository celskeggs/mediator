package dmi

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io/ioutil"
)

const pngHeader = "\x89PNG\r\n\x1a\n"

func readChunk(png []byte) (ctype string, data []byte, rest []byte, err error) {
	if len(png) < 4 {
		return "", nil, nil, errors.New("truncated PNG chunk")
	}
	length := binary.BigEndian.Uint32(png[0:4])
	if uint32(len(png)) < length+12 {
		return "", nil, nil, errors.New("truncated PNG chunk after length")
	}
	if crc32.ChecksumIEEE(png[4:length+8]) != binary.BigEndian.Uint32(png[length+8:length+12]) {
		return "", nil, nil, errors.New("crc32 mismatch")
	}
	return string(png[4:8]), png[8: 8+length], png[length+12:], nil
}

func decompress(zdata []byte) ([]byte, error) {
	rc, err := zlib.NewReader(bytes.NewReader(zdata))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(rc)
}

func parseDescription(png []byte) (string, error) {
	if string(png[:8]) != pngHeader {
		return "", errors.New("invalid PNG header")
	}
	png = png[8:]

	for len(png) > 0 {
		ctype, data, rest, err := readChunk(png)
		if err != nil {
			return "", err
		}
		if ctype == "zTXt" {
			if !bytes.Equal(data[:13], []byte("Description\000\000")) {
				continue
			}
			text, err := decompress(data[13:])
			if err != nil {
				return "", err
			}
			return string(text), nil
		}
		png = rest
	}
	return "", errors.New("did not find any description zTXt chunk")
}

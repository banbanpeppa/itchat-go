package util

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func getTimestamp() int {
	return int(time.Now().UnixNano() / 1000000)
}

func GetTimestamp() string {
	return strconv.Itoa(getTimestamp())
}

func GetR() string {
	return strconv.Itoa(-getTimestamp() / 1579)
}

func SleepSec(sec int) {
	time.Sleep(time.Duration(sec) * time.Second)
}

func GetRandomID(n int) string {
	rand.Seed(time.Now().Unix())
	return "e" + strconv.FormatFloat(rand.Float64(), 'f', n, 64)[2:]
}

func Store(data interface{}, filename string) error {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, buffer.Bytes(), 0600)
	if err != nil {
		return err
	}
	return nil
}

func Load(data interface{}, filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

func IsDirExist(path string) bool {
	p, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return p.IsDir()
	}
}

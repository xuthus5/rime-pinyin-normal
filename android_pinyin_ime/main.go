package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	f, err := os.OpenFile("rawdict_utf16_65105_freq.txt", os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(f)

	var string2Int = func(s string) int {
		f, err := strconv.ParseFloat(s, 10)
		if err != nil {
			fmt.Println(err)
			return 0
		}
		return int(math.Floor((f) + 0.5))
	}

	var buffer = []byte("---\nname: android_pinyin_simple\nversion: \"0.1\"\nsort: by_weight\n...\n")
	for {
		body, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		arr := strings.Split(string(body), " ")
		line := fmt.Sprintf("%s %s %d\n", arr[0], strings.Join(arr[3:], " "), string2Int(arr[1]))
		buffer = append(buffer, []byte(line)...)
	}

	ioutil.WriteFile("android_pinyin_simple.dict.yaml", buffer, fs.ModePerm)
}

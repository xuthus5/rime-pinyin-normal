package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/elliotchance/pie/pie"
	"github.com/spf13/cobra"
)

type WordsInfo struct {
	Pinyin string // 词组的拼音
	Words  string // 词组的汉字
	Weight int64  // 权重
}

var mergeDict *cobra.Command

func init() {
	var output string
	var inputs []string
	var weights []int64
	var example = "merge_dict -o custom_pinyin.dict.yaml -i \"linux.userdb.txt,android.userdb.txt,windows.userdb.txt\" -w 4,3,1"
	mergeDict = &cobra.Command{
		Use:     "merge_dict",
		Short:   "一个用来合并 rime 词典的程序",
		Example: example,
		Run: func(cmd *cobra.Command, args []string) {
			if len(inputs) == 0 {
				fmt.Println("please input user dict")
				return
			}

			if output == "" {
				fmt.Println("output empty!")
				return
			}

			if len(weights) != len(inputs) {
				fmt.Println("[input|weight] no match!")
				return
			}

			var dictM = make(map[string]*WordsInfo)

			for idx, input := range inputs {
				var w = weights[idx]
				m := readInputFile(input, w)

				for words, info := range m {
					_, exist := dictM[words]
					if exist {
						dictM[words].Pinyin = info.Pinyin
						dictM[words].Words = info.Words
						dictM[words].Weight += info.Weight
					} else {
						dictM[words] = &WordsInfo{
							Pinyin: info.Pinyin,
							Words:  info.Words,
							Weight: info.Weight,
						}
					}
				}

				fmt.Printf("%s merging count: %d\n", input, len(m))
			}

			// 重新排序写入
			var buffer []byte
			var sorts []string
			for _, info := range dictM {
				line := fmt.Sprintf("%s	%s	%d\n", info.Words, info.Pinyin, info.Weight)
				sorts = append(sorts, line)
			}
			// 重新排序一下
			sorts = pie.Strings(sorts).Unique().Sort()
			for _, line := range sorts {
				buffer = append(buffer, []byte(line)...)
			}
			ioutil.WriteFile(output, buffer, fs.ModePerm)
		},
	}

	mergeDict.Flags().StringVarP(&output, "output", "o", "", "合并后的字典输出位置(建议输出文件格式: a.dict.yaml)")
	mergeDict.Flags().StringSliceVarP(&inputs, "input", "i", []string{}, "合并字典来源")
	mergeDict.Flags().Int64SliceVarP(&weights, "weight", "w", []int64{}, "字典合并计算权重")
}

func readInputFile(src string, weight int64) map[string]WordsInfo {
	f, err := os.OpenFile(src, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(f)
	var m = make(map[string]WordsInfo)
	for {
		lineBody, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		line := string(lineBody)
		if strings.HasPrefix(line, "#") {
			continue
		}

		stab := strings.Split(line, "	")
		if len(stab) != 3 {
			continue
		}

		var traw = string2Int64(stab[2])

		var wi = WordsInfo{
			Words:  stab[0],
			Pinyin: stab[1],
			Weight: traw,
		}

		m[stab[0]] = wi
	}
	return m
}

func string2Int64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return i
}

func main() {
	mergeDict.Execute()
}

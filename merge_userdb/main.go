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

var mergeUserDB *cobra.Command

func init() {
	var output string
	var inputs []string
	var weights []int64
	mergeUserDB = &cobra.Command{
		Use:     "merge_userdb",
		Short:   "一个用来合并 rime 词典 userdb.txt 的程序",
		Example: "merge_userdb -o custom_pinyin.userdb.txt -i \"linux.userdb.txt,android.userdb.txt,windows.userdb.txt\" -w 4,3,1",
		Run: func(cmd *cobra.Command, args []string) {
			if len(inputs) == 0 {
				fmt.Println("please input userdb.txt")
				return
			}

			var dictM = make(map[string]*WordsInfo)

			sub := len(inputs) - len(weights)
			if sub > 0 {
				weights = append(weights, makeInt64OneSlice(sub)...)
			}

			for idx, input := range inputs {
				var w float64 = float64(weights[idx])
				m := readInputFile(input, w)

				for words, info := range m {
					dict, exist := dictM[words]
					if exist {
						dictM[words].Pinyin = info.Pinyin
						dictM[words].Words = info.Words
						dictM[words].C += info.C
						dictM[words].D += info.D
						dictM[words].T = (dict.T + info.T) / 2
					} else {
						dictM[words] = &WordsInfo{
							Pinyin: info.Pinyin,
							Words:  info.Words,
							C:      info.C,
							D:      info.D,
							T:      info.T,
						}
					}
				}

			}

			// 重新排序写入
			var buffer []byte
			var sorts []string
			for words, info := range dictM {
				line := fmt.Sprintf("%s	%s	c=%d d=%.4f t=%d\n", info.Pinyin, words, info.C, info.D, info.T)
				sorts = append(sorts, line)
			}

			sorts = pie.Strings(sorts).Unique().Sort()

			for _, line := range sorts {
				buffer = append(buffer, []byte(line)...)
			}

			ioutil.WriteFile(output, buffer, fs.ModePerm)

		},
	}

	mergeUserDB.Flags().StringVarP(&output, "output", "o", "custom.userdb.txt", "合并后的字典输出位置")
	mergeUserDB.Flags().StringSliceVarP(&inputs, "input", "i", []string{}, "合并字典来源")
	mergeUserDB.Flags().Int64SliceVarP(&weights, "weight", "w", []int64{}, "字典合并计算权重, 默认1")
}

func main() {
	mergeUserDB.Execute()
}

type WordsInfo struct {
	Pinyin string  // 词组的拼音
	Words  string  // 词组的汉字
	C      int64   // 输入法 commit 的次数 1，这个数可能因为输入时删除掉前面的词而减少，或者如果用户手动 shift+delete 删除掉候选词也会 reset 成 0
	D      float64 // 权重，结合时间，综合计算一个权重，随着时间推移，d 权重会衰减
	T      int64   // 时间，记录该候选词最近一次的时间
}

func readInputFile(src string, weight float64) map[string]WordsInfo {
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

		cdt := strings.Split(stab[2], " ")
		if len(cdt) != 3 {
			continue
		}

		var craw = string2Int64(splitEqualSymbol(cdt[0]))
		var draw = string2Float64(splitEqualSymbol(cdt[1]))
		var traw = string2Int64(splitEqualSymbol(cdt[2]))

		if draw <= 0.0001 {
			draw = 0.0001
		}

		var wi = WordsInfo{
			Pinyin: stab[0],
			Words:  stab[1],
			C:      craw,
			D:      draw * weight,
			T:      traw,
		}

		m[stab[1]] = wi
	}
	return m
}

func splitEqualSymbol(s string) string {
	arr := strings.Split(s, "=")
	if len(arr) == 2 {
		return arr[1]
	}
	return ""
}

func string2Int64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return i
}

func makeInt64OneSlice(l int) []int64 {
	var res []int64
	for i := 0; i < l; i++ {
		res = append(res, 1)
	}
	return res
}

func sumWeight(w []int64) int64 {
	var res int64
	for _, v := range w {
		res += v
	}
	return res
}

func string2Float64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return f
}

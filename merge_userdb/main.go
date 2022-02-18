package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/pie/pie"
	"github.com/spf13/cobra"
)

var mergeUserDB *cobra.Command

func init() {
	var output string
	var export string
	var inputs []string
	var weights []int64
	var example = fmt.Sprintf("merge_userdb -o custom_pinyin.userdb.txt -i \"linux.userdb.txt,android.userdb.txt,windows.userdb.txt\" -w 4,3,1" +
		"\nmerge_userdb -e universal.dict.yaml -i \"linux.user.db.txt,android.userdb.txt\" -w 4,2,1")
	mergeUserDB = &cobra.Command{
		Use:     "merge_userdb",
		Short:   "一个用来合并 rime 词典 userdb.txt 的程序",
		Example: example,
		Run: func(cmd *cobra.Command, args []string) {
			if len(inputs) == 0 {
				fmt.Println("please input userdb.txt")
				return
			}

			if output == "" && export == "" {
				fmt.Println("[output|export] need one")
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

			var dictBuffer []byte
			var dictSorts []string
			for words, info := range dictM {
				if output != "" {
					line := fmt.Sprintf("%s	%s	c=%d d=%.4f t=%d\n", info.Pinyin, words, info.C, info.D, info.T)
					sorts = append(sorts, line)
				}

				if export != "" {
					line := fmt.Sprintf("%s	%s	%d\n", info.Pinyin, words, info.C*1000)
					dictSorts = append(dictSorts, line)
				}
			}

			// 重新排序一下
			if output != "" {
				sorts = pie.Strings(sorts).Unique().Sort()
			}
			if export != "" {
				dictSorts = pie.Strings(dictSorts).Unique().Sort()
			}

			if output != "" {
				for _, line := range sorts {
					buffer = append(buffer, []byte(line)...)
				}
			}
			if export != "" {
				// 先写一下词典配置
				dictConfig := fmt.Sprintf("---\nname: %s\nversion: \"%s\"\nsort: by_weight\nuse_preset_vocabulary: false\n...\n", getBaseName(export), getToday())
				dictBuffer = append(dictBuffer, []byte(dictConfig)...)
				for _, line := range dictSorts {
					dictBuffer = append(dictBuffer, []byte(line)...)
				}
			}

			if output != "" {
				ioutil.WriteFile(output, buffer, fs.ModePerm)
			}
			if export != "" {
				ioutil.WriteFile(export, dictBuffer, fs.ModePerm)
			}
		},
	}

	mergeUserDB.Flags().StringVarP(&output, "output", "o", "", "合并后的字典快照输出位置(建议输出文件格式: a.userdb.txt)")
	mergeUserDB.Flags().StringSliceVarP(&inputs, "input", "i", []string{}, "合并字典快照来源")
	mergeUserDB.Flags().Int64SliceVarP(&weights, "weight", "w", []int64{}, "字典快照合并计算权重, 默认1")
	mergeUserDB.Flags().StringVarP(&export, "export", "e", "", "词典导出位置.合并后依据数据导出一份真实的词典(建议输出文件格式: a.dict.yaml)")
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

func string2Float64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return f
}

func getToday() string {
	tm := time.Now()
	return tm.Format("2006.01.02")
}

func getBaseName(src string) string {
	p := filepath.Base(src)
	if p == "." {
		return "universal"
	}
	if strings.HasSuffix(p, ".dict.yaml") {
		s := strings.Split(p, ".")
		return s[0]
	}
	return p
}

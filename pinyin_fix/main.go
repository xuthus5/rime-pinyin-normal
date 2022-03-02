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
	"github.com/mozillazg/go-pinyin"
	"github.com/spf13/cobra"
)

type WordsInfo struct {
	Pinyin []string // 词组的拼音
	Hans   string   // 词组的汉字
	Weight int64    // 词组的权重
}

var pinyinFix *cobra.Command
var pyArgs = pinyin.NewArgs()

func init() {
	var name string
	var output string
	var input string
	var example = ""
	pinyinFix = &cobra.Command{
		Use:     "pinyin_fix",
		Short:   "一个用来修复 rime 词典错误拼音的程序",
		Example: example,
		Run: func(cmd *cobra.Command, args []string) {
			if input == "" || output == "" {
				fmt.Println("[input|output] missing arguments")
				return
			}
			var multiDict = make(map[string]*WordsInfo)
			var singleDict = make(map[string]*WordsInfo)
			m := readInputFile(input)
			for words, info := range m {
				if len([]rune(words)) != 1 {
					multiDict[words] = &WordsInfo{
						Pinyin: info.Pinyin,
						Hans:   info.Hans,
						Weight: info.Weight,
					}
					continue
				}
				p := getPinyin(words)
				singleDict[words] = &WordsInfo{
					Hans:   info.Hans,
					Weight: info.Weight,
					Pinyin: p,
				}
			}

			// 重新排序写入词组字典
			{
				var buffer []byte
				var sorts []string
				for words, info := range multiDict {
					for _, p := range info.Pinyin {
						line := fmt.Sprintf("%s	%s	%d\n", words, p, info.Weight)
						sorts = append(sorts, line)
					}
				}
				buffer = append(buffer, getDictNameDecls(name)...)
				// 重新排序一下
				sorts = pie.Strings(sorts).Unique().Sort()
				for _, line := range sorts {
					buffer = append(buffer, []byte(line)...)
				}
				ioutil.WriteFile(fmt.Sprintf("%s/%s.dict.yaml", output, name), buffer, fs.ModePerm)
			}

			// 重新排序写入汉字
			{
				var buffer []byte
				var sorts []string
				for words, info := range singleDict {
					for _, p := range info.Pinyin {
						line := fmt.Sprintf("%s	%s	%d\n", words, p, info.Weight)
						sorts = append(sorts, line)
					}
				}
				buffer = append(buffer, getDictNameDecls(fmt.Sprintf("%s.single", name))...)
				// 重新排序一下
				sorts = pie.Strings(sorts).Unique().Sort()
				for _, line := range sorts {
					buffer = append(buffer, []byte(line)...)
				}
				ioutil.WriteFile(fmt.Sprintf("%s/%s.single.dict.yaml", output, name), buffer, fs.ModePerm)
			}
		},
	}

	pinyinFix.Flags().StringVarP(&output, "output", "o", "", "字典拼音修复后的输出位置")
	pinyinFix.Flags().StringVarP(&input, "input", "i", "", "需要修复的字典位置")
	pinyinFix.Flags().StringVarP(&name, "name", "n", "base", "字典名(默认生成base.dict.yaml)")
}

func main() {
	pinyinFix.Execute()
}

func getPinyin(hans string) []string {
	pyArgs.Heteronym = true
	res := pinyin.Pinyin(hans, pyArgs)
	return pie.Strings(res[0]).Unique()
}

func readInputFile(src string) map[string]WordsInfo {
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

		var wi = WordsInfo{
			Pinyin: []string{stab[1]},
			Hans:   strings.Trim(stab[0], " "),
			Weight: string2Int64(stab[2]),
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

func getDictNameDecls(name string) []byte {
	var lines []byte
	lines = append(lines, []byte("---\n")...)
	lines = append(lines, []byte(fmt.Sprintf("name: %s\n", name))...)
	lines = append(lines, []byte("version: \"0.1\"\n")...)
	lines = append(lines, []byte("sort: by_weight\n")...)
	lines = append(lines, []byte("...\n")...)
	return lines
}

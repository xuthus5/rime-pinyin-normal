# MERGE_USERDB

是一个用来合并 rime 的 userdb.txt 的处理程序。

你可以直接通过 `go install github.com/xuthus5/rime-pinyin-normal/merge_userdb@latest` 进行安装

## 使用说明

```shell
> merge_userdb -h
一个用来合并 rime 词典 userdb.txt 的程序

Usage:
  merge_userdb [flags]

Examples:
merge_userdb -o custom_pinyin.userdb.txt -i "linux.userdb.txt,android.userdb.txt,windows.userdb.txt" -w 4,3,1

Flags:
  -h, --help                help for merge_userdb
  -i, --input strings       合并字典来源
  -o, --output string       合并后的字典输出位置 (default "custom.userdb.txt")
  -w, --weight int64Slice   字典合并计算权重, 默认1 (default [])
```
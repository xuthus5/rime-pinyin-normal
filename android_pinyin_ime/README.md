此处是一个基于Go的安卓字典转换服务，他会将其格式转为rime的字典

**注：** `rawdict_utf16_65105_freq.txt` 已被编码为 `UTF-8` 无需再进行编码处理。

字典路径: [https://android.googlesource.com/platform/packages/inputmethods/PinyinIME/+/refs/heads/master/jni/data/rawdict_utf16_65105_freq.txt](https://android.googlesource.com/platform/packages/inputmethods/PinyinIME/+/refs/heads/master/jni/data/rawdict_utf16_65105_freq.txt)

字典仓库下载：[https://android.googlesource.com/platform/packages/inputmethods/PinyinIME/+archive/refs/heads/master.tar.gz](https://android.googlesource.com/platform/packages/inputmethods/PinyinIME/+archive/refs/heads/master.tar.gz)

## 转换

```shell
go run main.go
```
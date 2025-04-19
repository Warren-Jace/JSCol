# JSCol

## 简介

JSCol 是一个用 Go 开发的命令行工具，支持多线程并发调用多种 JS 收集工具（如 URLFinder、getJS、jsfinder、subjs），自动收集指定 URL 下的 JS 文件地址。

## 功能特性

- 支持并发调用多个 JS 工具，易于扩展
- 详细日志与错误处理
- 代码结构清晰，便于维护
- 遵循 OWASP 安全最佳实践，避免数据外泄

## 使用方法

1.  安装依赖的外部工具（如 URLFinder、getJS、jsfinder、subjs），并将它们所在的目录添加到系统环境变量 `$PATH` 中。
2.  安装 Go 1.18+。
3.  编译：

    ```bash
    go build -o jscol main.go
    ```

4.  运行：

    ```bash
    ./jscol -urlfile urls.txt -config config.json
    ```

    *   `-urlfile`:  指定包含目标 URL 的文件。
    *   `-config`: 指定配置文件的路径 (默认为 `config.json`)。

5.  收集结果将保存在各个工具指定的输出文件中，以及日志中。

## 目录结构
```bash
.
├── config.json       # 配置文件
├── go.mod            # Go Modules 文件
├── main.go           # 主程序入口
├── README.md         # 说明文档
├── collector/        # 收集器目录
│   ├── collector.go  # 收集器接口定义
│   ├── getjs.go      # getJS 收集器实现
│   ├── jsfinder.go   # jsfinder 收集器实现
│   ├── subjs.go      # subjs 收集器实现
│   └── urlfinder.go  # URLFinder 收集器实现
├── results/          # 结果输出目录
│   ├── getJS.txt # getJS 输出
│   ├── subjs.txt               # subjs 输出
│   ├── jsfinder.txt       # jsfinder 输出
│   └── URLFinder/              # URLFinder 输出目录
│       └── ...
├── tools/            # 工具目录
│   ├── getJS.exe
│   ├── jsfinder.exe
│   ├── subjs.exe
│   └── URLFinder.exe
└── utils/            # 工具函数目录
    ├── logger.go     # 日志
    └── tool_manager.go # 工具管理
```
## 配置文件 (config.json)

配置文件用于指定各个工具的路径和参数。 示例：

```json
{
  "jsfinder": {
    "Name": "jsfinder",
    "Path": "tools/jsfinder.exe",
    "Args": ["-l", "urls.txt", "-c", "50", "-s", "-o", "results/urls-jsfinder.txt"]
  },
  "urlfinder": {
    "Name": "URLFinder",
    "Path": "tools/URLFinder.exe",
    "Args": ["-s", "all", "-m", "3", "-f", "urls.txt", "-o", "results/URLFinder/"]
  },
  "getjs": {
    "Name": "getJS",
    "Path": "tools/getJS.exe",
    "Args": ["-input", "urls.txt", "-complete", "-resolve", "-output", "results/hosts-getJS-results.txt"]
  },
  "subjs": {
    "Name": "subjs",
    "Path": "tools/subjs.exe",
    "Args": ["-i", "urls.txt"]
  }
}
```
## 扩展说明
如需集成新的 JS 工具，只需在 collector 目录下新建实现 JSCollector 接口的新文件，并在 main.go 的 tools 切片中添加即可。

## 安全声明
本工具不会上传或外泄任何用户数据
所有操作均在本地执行
建议在合法授权范围内使用

## 依赖
Go 1.18+
URLFinder
getJS
jsfinder
subjs

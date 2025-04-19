// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"jscol/collector"
	"jscol/utils"
)

// 结果结构体
type JSResult struct {
	Tool   string   `json:"tool"`
	JSURLs []string `json:"js_urls"`
	Error  string   `json:"error,omitempty"`
}

// ToolConfig 结构体用于存储工具的配置信息
type ToolConfig struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
}

// Config 结构体用于存储配置信息
type Config struct {
	Gau       ToolConfig `json:"gau"`
	JSFinder  ToolConfig `json:"jsfinder"`
	URLFinder ToolConfig `json:"urlfinder"`
	GetJS     ToolConfig `json:"getjs"`
	SubJS     ToolConfig `json:"subjs"`
}

// Stringer interface
type Stringer interface {
	String() string
}

func main() {
	utils.InitLogger()

	// 检查并创建 results 目录
	resultsDir := "results"
	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(resultsDir, 0755); err != nil {
			log.Fatalf("创建 results 目录失败: %v", err)
		}
	}

	// 检查并准备工具
	if err := utils.CheckAndPrepareTools(); err != nil {
		log.Fatalf("工具检查或准备失败: %v", err)
	}
	// urlPtr 现在是文件名
	urlFilePtr := flag.String("urlfile", "", "包含目标URL的文件")
	configPtr := flag.String("config", "config.json", "配置文件名") // 新增配置文件参数
	flag.Parse()

	if *urlFilePtr == "" {
		log.Fatal("必须指定包含目标URL的文件，例如：-urlfile urls.txt")
	}

	// 加载配置文件
	var config Config
	configFile, err := os.Open(*configPtr)
	if err != nil {
		log.Fatalf("打开配置文件失败: %v", err)
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	tools := []collector.JSCollector{
		collector.JSFinderCollector{
			ToolPath: config.JSFinder.Path,
			Args:     config.JSFinder.Args,
		},
		collector.URLFinderCollector{
			ToolPath: config.URLFinder.Path,
			Args:     config.URLFinder.Args,
		},
		collector.GetJSCollector{
			ToolPath: config.GetJS.Path,
			Args:     config.GetJS.Args,
		},
		collector.SubJSCollector{
			ToolPath: config.SubJS.Path,
			Args:     config.SubJS.Args,
		},
		// 后续可在这里添加更多Collector
	}

	var wg sync.WaitGroup
	resultChan := make(chan JSResult, len(tools))

	for _, t := range tools {
		wg.Add(1)
		go func(tc collector.JSCollector) {
			defer wg.Done()
			utils.Logger.Printf("开始用%s收集JS...", tc.Name())
			var jsUrls []string
			var err error

			if tc.Name() == "subjs" {
				// 特殊处理 subjs
				cmd := exec.Command(tc.(collector.SubJSCollector).ToolPath, tc.(collector.SubJSCollector).Args...)
				// 将 urls.txt 作为输入传递给 subjs
				// Open urls.txt and pass it as stdin
				file, err := os.Open(*urlFilePtr)
				if err != nil {
					utils.Logger.Printf("[%s] 发生错误: %v", tc.Name(), err)
					return
				}
				defer file.Close()
				cmd.Stdin = file

				// Print the command being executed
				commandStr := strings.Join(cmd.Args, " ")
				utils.Logger.Printf("执行命令: %s", commandStr)

				output, err := cmd.CombinedOutput()
				if err != nil {
					utils.Logger.Printf("[%s] 发生错误: %v", tc.Name(), err)
				} else {
					// 将结果写入文件
					subjsOutputFile := filepath.Join(resultsDir, "subjs.txt")
					f, err := os.OpenFile(subjsOutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						utils.Logger.Printf("[%s] 写入文件失败: %v", tc.Name(), err)
					}
					defer f.Close()
					if _, err := f.WriteString(string(output)); err != nil {
						utils.Logger.Printf("[%s] 写入文件失败: %v", tc.Name(), err)
					}
					// 从输出中提取 URL（如果需要）
					// 这里假设输出是每行一个 URL
					lines := strings.Split(string(output), "\n")
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line != "" {
							jsUrls = append(jsUrls, line)
						}
					}

				}
			} else {
				// Print the command being executed
				var commandStr string
				toolName := tc.(interface{ Name() string }).Name()
				switch toolName {
				case "jsfinder":
					commandStr = fmt.Sprintf("%s %s", toolName, strings.Join(config.JSFinder.Args, " "))
				case "URLFinder":
					commandStr = fmt.Sprintf("%s %s", toolName, strings.Join(config.URLFinder.Args, " "))
				case "getJS":
					commandStr = fmt.Sprintf("%s %s", toolName, strings.Join(config.GetJS.Args, " "))
				default:
					commandStr = fmt.Sprintf("Unknown tool: %s", toolName)
				}

				utils.Logger.Printf("执行命令: %s", commandStr)
				// Pass urlsFile to Collect method
				jsUrls, err = tc.Collect(*urlFilePtr)
			}

			res := JSResult{
				Tool:   tc.Name(),
				JSURLs: jsUrls,
			}
			if err != nil {
				utils.Logger.Printf("[%s] 发生错误: %v", tc.Name(), err)
				res.Error = err.Error()
			}
			resultChan <- res
		}(t)
	}

	wg.Wait()
	close(resultChan)

	utils.Logger.Printf("所有工具收集完成")
	fmt.Println("收集完成")
}

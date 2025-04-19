package collector

import (
	"bufio"
	"os/exec"
	"strings"
)

// SubJSCollector 调用subjs工具
type SubJSCollector struct {
	ToolPath string
	Args     []string
}

// Name 返回收集器的名称
func (s SubJSCollector) Name() string {
	return "subjs"
}

// Collect 使用subjs工具收集指定URL的JavaScript文件URL
func (s SubJSCollector) Collect(url string) ([]string, error) {
	var jsUrls []string

	cmd := exec.Command(s.ToolPath, s.Args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, ".js") {
			jsUrls = append(jsUrls, line)
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return jsUrls, nil
}

// collector/jsfinder.go
package collector

import (
	"bufio"
	"os/exec"
	"strings"
)

// JSFinderCollector 调用jsfinder工具
type JSFinderCollector struct {
	ToolPath string
	Args     []string
}

// Name 返回收集器的名称
func (j JSFinderCollector) Name() string {
	return "jsfinder"
}

// Collect 使用jsfinder工具收集指定URL的JavaScript文件URL
func (j JSFinderCollector) Collect(url string) ([]string, error) {
	var jsUrls []string
	cmd := exec.Command(j.ToolPath, j.Args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "http") && strings.HasSuffix(line, ".js") {
			jsUrls = append(jsUrls, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return jsUrls, nil
}

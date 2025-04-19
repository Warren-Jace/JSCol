package collector

import (
	"bufio"
	"os/exec"
	"strings"
)

// URLFinderCollector 调用URLFinder工具
type URLFinderCollector struct {
	ToolPath string
	Args     []string
}

// Name 返回收集器的名称
func (u URLFinderCollector) Name() string {
	return "URLFinder"
}

// Collect 使用URLFinder工具收集指定URL的JavaScript文件URL
func (u URLFinderCollector) Collect(url string) ([]string, error) {
	var jsUrls []string

	cmd := exec.Command(u.ToolPath, u.Args...)
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

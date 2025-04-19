package collector

import (
	"bufio"
	"os/exec"
	"strings"
)

// GetJSCollector 调用getJS工具
type GetJSCollector struct {
	ToolPath string
	Args     []string
}

// Name 返回收集器的名称
func (g GetJSCollector) Name() string {
	return "getJS"
}

// Collect 使用getJS工具收集指定URL的JavaScript文件URL
func (g GetJSCollector) Collect(url string) ([]string, error) {
	var jsUrls []string

	cmd := exec.Command(g.ToolPath, g.Args...)
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

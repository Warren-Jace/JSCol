package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ToolConfig 结构体用于存储工具的配置信息
type ToolConfig struct {
	Name string   `json:"Name"`
	Path string   `json:"Path"`
	Args []string `json:"Args"`
}

// Config 结构体用于存储配置信息
type Config struct {
	JSFinder  ToolConfig `json:"jsfinder"`
	URLFinder ToolConfig `json:"urlfinder"`
	GetJS     ToolConfig `json:"getjs"`
	SubJS     ToolConfig `json:"subjs"`
}

// CheckAndPrepareTools 检查并准备所有工具
func CheckAndPrepareTools() error {
	toolsDir := "tools"
	if _, err := os.Stat(toolsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(toolsDir, 0755); err != nil {
			return fmt.Errorf("创建tools目录失败: %v", err)
		}
	}

	// 加载配置文件
	config, err := loadConfig("config.json")
	if err != nil {
		return err
	}

	osType := runtime.GOOS

	tools := []ToolConfig{config.JSFinder, config.URLFinder, config.GetJS, config.SubJS}
	for _, toolConfig := range tools {
		if err := prepareTool(toolConfig, osType); err != nil {
			return err
		}
	}

	return nil
}

func loadConfig(configFile string) (*Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("打开配置文件失败: %v", err)
	}
	defer f.Close()

	var config Config
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}
	return &config, nil
}

// prepareTool 检查并准备指定的工具
func prepareTool(toolConfig ToolConfig, osType string) error {
	toolPath := toolConfig.Path
	toolName := toolConfig.Name

	var executableName string
	if osType == "windows" {
		executableName = toolPath + ".exe"
	} else {
		executableName = toolPath
	}

	if _, err := os.Stat(executableName); os.IsNotExist(err) {
		fmt.Printf("%s工具不存在，正在下载...\n", toolName)

		downloadURL, needsGoBuild, err := getDownloadInfo(toolName, osType)
		if err != nil {
			return err
		}

		if needsGoBuild {
			if err := downloadAndBuild(downloadURL, toolName, executableName); err != nil {
				return err
			}
		} else if strings.Contains(downloadURL, ".zip") || strings.Contains(downloadURL, ".tar.gz") {
			if err := downloadAndExtract(downloadURL, toolName, osType, executableName); err != nil {
				return err
			}
		} else {
			if err := downloadFile(downloadURL, executableName); err != nil {
				return fmt.Errorf("下载%s工具失败: %v", toolName, err)
			}
		}

		if osType != "windows" {
			if err := os.Chmod(executableName, 0755); err != nil {
				return fmt.Errorf("设置%s工具执行权限失败: %v", toolName, err)
			}
		}
		fmt.Printf("%s工具下载完成\n", toolName)
	}
	return nil
}

func getDownloadInfo(toolName string, osType string) (string, bool, error) {
	switch toolName {
	case "URLFinder":
		if osType == "windows" {
			return "https://github.com/pingc0y/URLFinder/releases/download/2023.9.9/URLFinder_Windows_x86_64.zip", false, nil
		} else if osType == "linux" {
			return "https://github.com/pingc0y/URLFinder/releases/download/2023.9.9/URLFinder_Linux_x86_64.tar.gz", false, nil
		} else if osType == "darwin" {
			return "https://github.com/pingc0y/URLFinder/releases/download/2023.9.9/URLFinder_Darwin_x86_64.tar.gz", false, nil
		}
		return "", false, fmt.Errorf("不支持的操作系统: %s", osType)
	case "getJS":
		return "https://github.com/003random/getJS", true, nil // 需要 go build
	case "subjs":
		return "https://github.com/lc/subjs", true, nil // 需要 go build
	case "jsfinder":
		return "https://github.com/kacakb/jsfinder", true, nil // 需要 go build
	default:
		return "", false, fmt.Errorf("不支持的工具: %s", toolName)
	}
}

func downloadAndBuild(downloadURL string, toolName string, executableName string) error {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "jscol_tmp")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir) // 确保在函数退出时删除临时目录

	// 下载源代码
	archivePath := filepath.Join(tmpDir, toolName+".git")
	cmd := exec.Command("git", "clone", downloadURL, archivePath)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("克隆%s仓库失败: %v, output: %s", toolName, err, string(output))
	}

	// 构建可执行文件
	buildCmd := exec.Command("go", "build", "-o", executableName, ".")
	buildCmd.Dir = archivePath
	output, err = buildCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("构建%s工具失败: %v, output: %s", toolName, err, string(output))
	}

	// 移动可执行文件
	if err := os.Rename(filepath.Join(archivePath, executableName), executableName); err != nil {
		return fmt.Errorf("移动%s工具失败: %v", toolName, err)
	}

	return nil
}

func downloadAndExtract(downloadURL string, toolName string, osType string, executableName string) error {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "jscol_tmp")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir) // 确保在函数退出时删除临时目录

	archivePath := filepath.Join(tmpDir, getArchiveName(downloadURL))

	if err := downloadFile(downloadURL, archivePath); err != nil {
		return fmt.Errorf("下载%s工具失败: %v", toolName, err)
	}

	// 解压文件
	extractDir := filepath.Join(tmpDir, toolName+"_extracted")
	if err := extractArchive(archivePath, extractDir); err != nil {
		return fmt.Errorf("解压%s工具失败: %v", toolName, err)
	}

	// 查找解压后的可执行文件
	executablePath, err := findExecutable(extractDir, toolName, osType)
	if err != nil {
		return err
	}

	// 移动并重命名可执行文件
	if err := os.Rename(executablePath, executableName); err != nil {
		return fmt.Errorf("移动%s工具失败: %v", toolName, err)
	}

	return nil
}

func getArchiveName(downloadURL string) string {
	parts := strings.Split(downloadURL, "/")
	return parts[len(parts)-1]
}

func extractArchive(archivePath string, extractDir string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return unzip(archivePath, extractDir)
	} else if strings.HasSuffix(archivePath, ".tar.gz") {
		return untar(archivePath, extractDir)
	}
	return fmt.Errorf("不支持的压缩文件格式: %s", archivePath)
}

func findExecutable(extractDir string, toolName string, osType string) (string, error) {
	executableName := toolName
	if osType == "windows" {
		executableName += ".exe"
	}
	var executablePath string
	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(path, executableName) {
			executablePath = path
			return io.EOF // 找到可执行文件，停止遍历
		}
		return nil
	})
	if err == io.EOF {
		return executablePath, nil
	}
	if executablePath == "" {
		return "", fmt.Errorf("未找到%s工具的可执行文件", toolName)
	}
	return executablePath, err
}

// downloadFile 下载文件到指定路径
func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载文件失败: %s, 状态码: %d", url, resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// unzip 解压 zip 文件
func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.Create(fpath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}
	return nil
}

// untar 解压 tar.gz 文件
func untar(src string, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
		}
	}
}

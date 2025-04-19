// collector/collector.go
package collector

// JSCollector接口，便于扩展
type JSCollector interface {
	Name() string
	Collect(url string) ([]string, error)
}

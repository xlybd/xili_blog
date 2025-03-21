package utils

import (
	"io/fs"
	"os"
	"server/global"

	"gopkg.in/yaml.v3"
)

const configFile = "config.yaml"

// 获取配置文件路径，支持环境变量
func getConfigFilePath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return configFile
}

// LoadYAML 从文件中读取 YAML 数据并返回字节数组
func LoadYAML() ([]byte, error) {
	configFile := getConfigFilePath()
	return os.ReadFile(configFile)
}

// SaveYAML 将全局配置对象保存为 YAML 格式到文件
func SaveYAML() error {
	configFile := getConfigFilePath()
	byteData, err := yaml.Marshal(global.Config)
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, byteData, fs.ModePerm)
}

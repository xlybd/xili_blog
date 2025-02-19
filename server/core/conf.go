package core

import (
	"log"
	"server/config"
	"server/utils"

	"gopkg.in/yaml.v3"
)

// InitConf 从 YAML 文件加载配置
func InitConf() *config.Config {
	c := &config.Config{}
	yamlConf, err := utils.LoadYAML()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if err = yaml.Unmarshal(yamlConf, c); err != nil {
		log.Fatalf("Failed to umarshal YAML configuration: %v", err)
	}
	return c
}

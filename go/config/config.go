package config

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "minioDeploy/global"
    "os"
)

type Config struct {
    SystemUser        string `yaml:"systemUser"`
    Data              string `yaml:"data"`
    Opts              string `yaml:"opts"`
    MinIORootUser     string `yaml:"minioRootUser"`
    MinIORootPassword string `yaml:"minioRootPassword"`
    Address           string `yaml:"address"`
    ConsoleAddress    string `yaml:"consoleAddress"`
    Region            string `yaml:"region"`
}

func GetConfig() *Config {
    file, err := os.Open(global.ConfigFilePath)
    if err != nil {
        fmt.Printf("Unable to open file: %v\n", err)
        return nil
    }
    defer func(file *os.File) {
        _ = file.Close()
    }(file)
    config := &Config{}
    decoder := yaml.NewDecoder(file)
    if err := decoder.Decode(config); err != nil {
        fmt.Printf("Failed to parse YAML file: %v\n", err)
        return nil
    }
    if config.SystemUser == "" {
        config.SystemUser = "minio-user"
    }
    if config.MinIORootUser == "" {
        config.MinIORootUser = "superadmin"
    }
    if config.MinIORootPassword == "" {
        config.MinIORootPassword = "superadmin"
    }
    if config.Data == "" {
        config.Data = "data"
    }
    if config.Address == "" {
        config.Address = ":9000"
    }
    if config.ConsoleAddress == "" {
        config.ConsoleAddress = ":9001"
    }
    return config
}

package util

import "os"

func PathExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil || !os.IsNotExist(err)
}

func IsFile(path string) bool {
    stat, err := os.Stat(path)
    return err == nil && !stat.IsDir()
}

func IsDir(path string) bool {
    stat, err := os.Stat(path)
    return err == nil && stat.IsDir()
}

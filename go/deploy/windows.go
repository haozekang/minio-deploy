//go:build windows

package deploy

import (
    "bufio"
    "fmt"
    config2 "minioDeploy/config"
    "os"
    "os/exec"
    "path/filepath"
)

func Deploy(exePath string, config *config2.Config) bool {
    exeAbsPath, err := filepath.Abs(exePath)
    if err != nil {
        return false
    }
    dataAbsPath, err := filepath.Abs(config.Data)
    if err != nil {
        return false
    }
    fmt.Println("Minio ExePath :", exeAbsPath)
    fmt.Println("Minio DataPath:", dataAbsPath)
    _ = os.Setenv("MINIO_ROOT_USER", config.MinIORootUser)
    _ = os.Setenv("MINIO_ROOT_PASSWORD", config.MinIORootPassword)
    fmt.Println("Minio Root User    :", config.MinIORootUser)
    fmt.Println("Minio Root Password:", config.MinIORootPassword)
    cmd := exec.Command(
        `D:\GolandProjects\minioDeploy\minio.exe`,
        "server",
        config.Data,
        "--address="+config.Address,
        "--console-address="+config.ConsoleAddress,
    )

    stdoutPipe, _ := cmd.StdoutPipe()
    stderrPipe, _ := cmd.StderrPipe()

    err = cmd.Start()
    if err != nil {
        fmt.Println("Error Start Minio Service:", err)
        return false
    }

    // 分别异步读取输出
    go func() {
        scanner := bufio.NewScanner(stdoutPipe)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    }()

    go func() {
        scanner := bufio.NewScanner(stderrPipe)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    }()

    err = cmd.Wait()
    if err != nil {
        fmt.Println("MinIO Start Error：", err)
    }
    return true
}

package main

import (
    "fmt"
    "io"
    config2 "minioDeploy/config"
    "minioDeploy/deploy"
    "minioDeploy/global"
    "minioDeploy/util"
    "net/http"
    "os"
    "path/filepath"
    "runtime"
    "strconv"
    "time"
)

func main() {
    fmt.Printf("%s\nVersion: %s\t Author:%s\n", global.Title, global.Version, global.Author)
    osEnv := runtime.GOOS
    if osEnv != "windows" && osEnv != "darwin" && osEnv != "linux" {
        fmt.Println("System not supported!")
    }
    fmt.Println("System Env:", osEnv)
    url := ""
    filePath := ""
    switch osEnv {
    case "linux":
        url = "https://dl.min.io/server/minio/release/linux-amd64/minio"
        filePath = "minio"
        break
    case "windows":
        url = "https://dl.min.io/server/minio/release/windows-amd64/minio.exe"
        filePath = "minio.exe"
        break
    case "darwin":
        url = "https://dl.min.io/server/minio/release/darwin-arm64/minio"
        filePath = "minio"
        break
    }
    if url == "" {
        fmt.Println("No URL!")
        return
    }
    downloadFlag := false
    // 可执行文件不存在，则下载
    if (osEnv == "windows" && !util.PathExists(filePath)) || ((osEnv == "linux" || osEnv == "darwin") && !util.PathExists(`/usr/local/bin/minio`)) {
        downloadFlag = true
        req, _ := http.NewRequest("HEAD", url, nil)
        resp, err := http.DefaultClient.Do(req)
        var size int64
        size = -1
        if err == nil && resp.Header.Get("Content-Length") != "" {
            size, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
            fmt.Printf("[%s]File size:%d", filepath.Base(filePath), size)
        }
        resp, err = http.Get(url)
        if err != nil {
            fmt.Println("Download failed:", err)
            return
        }
        defer func(Body io.ReadCloser) {
            _ = Body.Close()
        }(resp.Body)
        if size == -1 {
            fmt.Println("Unable to get file size!")
        }
        out, err := os.Create(filePath)
        if err != nil {
            fmt.Println("Failed to create file:", err)
            return
        }
        defer func(out *os.File) {
            _ = out.Close()
        }(out)
        var downloaded int64 = 0
        buf := make([]byte, 32*1024)
        lastTime := time.Now()

        for {
            n, err := resp.Body.Read(buf)
            if n > 0 {
                _, writeErr := out.Write(buf[:n])
                if writeErr != nil {
                    fmt.Println("Failed to write file:", writeErr)
                    return
                }
                downloaded += int64(n)

                // 每 100ms 更新一次进度条
                if time.Since(lastTime) > 100*time.Millisecond {
                    printProgress(downloaded, size)
                    lastTime = time.Now()
                }
            }
            if err != nil {
                if err == io.EOF {
                    break
                }
                fmt.Println("Read Error:", err)
                return
            }
        }

        printProgress(downloaded, size)
        fmt.Println("\nDownload Complete!")
    }
    if downloadFlag {
        if !util.PathExists(filePath) {
            fmt.Println("File not found!")
            return
        }
    } else {
        if osEnv == "windows" {
            if !util.PathExists(filePath) {
                fmt.Println("File not found!")
                return
            }
        } else {
            if !util.PathExists(filePath) && !util.PathExists(`/usr/local/bin/minio`) {
                fmt.Println("File not found!")
                return
            }
        }
    }
    config := config2.GetConfig()
    if !deploy.Deploy(filePath, config) {
        fmt.Println("Deployment Failure!")
        return
    }
    fmt.Println("Deployment Completed!")
}

func printProgress(downloaded, total int64) {
    const barWidth = 50
    if total <= 0 {
        fmt.Printf("\rDownloaded [%d] byte", downloaded)
        return
    }
    percent := float64(downloaded) / float64(total)
    done := int(percent * barWidth)

    bar := "[" + repeat("█", done) + repeat("░", barWidth-done) + "]"
    fmt.Printf("\r%s %.2f%%", bar, percent*100)
}

func repeat(s string, count int) string {
    result := ""
    for i := 0; i < count; i++ {
        result += s
    }
    return result
}

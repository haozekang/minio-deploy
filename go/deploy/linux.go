//go:build linux

package deploy

import (
    "fmt"
    "minioDeploy/util"
    "os"
    "os/exec"
    "os/user"
    config2 "minioDeploy/config"
    "path/filepath"
)

func Deploy(exePath string, config *config2.Config) bool {
    var cmd = &exec.Cmd{}
    exeAbsPath, err := filepath.Abs(exePath)
    if err != nil {
        return false
    }
    dataAbsPath, err := filepath.Abs(config.Data)
    if err != nil {
        return false
    }

    currentUser, err := user.Current()
    if err != nil {
        fmt.Println("Unable to obtain current user information:", err)
        os.Exit(1)
    }

    if currentUser.Uid != "0" {
        fmt.Println("Please run this program as root user.")
        os.Exit(1)
    }

    fmt.Println("Minio ExePath :", exeAbsPath)
    fmt.Println("Minio DataPath:", dataAbsPath)
    _ = os.Setenv("MINIO_ROOT_USER", config.MinIORootUser)
    _ = os.Setenv("MINIO_ROOT_PASSWORD", config.MinIORootPassword)
    fmt.Println("Minio Root User    :", config.MinIORootUser)
    fmt.Println("Minio Root Password:", config.MinIORootPassword)

    if util.PathExists(exeAbsPath) {
        cmd = exec.Command("chmod", "+x", exeAbsPath)
        err = cmd.Run()
        if err != nil {
            fmt.Println("chmod error:", err)
            fmt.Println("cmd:", cmd.String())
            return false
        }
        cmd = exec.Command("mv", exePath, "/usr/local/bin/")
        err = cmd.Run()
        if err != nil {
            fmt.Println("mv error:", err)
            fmt.Println("cmd:", cmd.String())
            return false
        }
    }
    _, err = user.Lookup(config.SystemUser)
    if err != nil {
        cmd = exec.Command("useradd", "-r", config.SystemUser, "-s", "/sbin/nologin")
        err = cmd.Run()
        if err != nil {
            fmt.Println("useradd error:", err)
            fmt.Println("cmd:", cmd.String())
            return false
        }
    }
    if !util.PathExists(config.Data) {
        cmd = exec.Command("mkdir", config.Data)
        err = cmd.Run()
        if err != nil {
            fmt.Println("mkdir data error:", err)
            fmt.Println("cmd:", cmd.String())
            return false
        }
    }
    cmd = exec.Command("chown", "-R", config.SystemUser+":"+config.SystemUser, config.Data)
    err = cmd.Run()
    if err != nil {
        fmt.Println("chown data error:", err)
        fmt.Println("cmd:", cmd.String())
        return false
    }
    configFile, err := os.Create("/etc/default/minio")
    if err != nil {
        fmt.Println("create config file error:", err)
        return false
    }
    configFile.WriteString("MINIO_VOLUMES=\"" + config.Data + "\"\n" +
        "MINIO_OPTS=\"--address '" + config.Address + "' --console-address '" + config.ConsoleAddress + "'\"\n" +
        "MINIO_ACCESS_KEY=\"" + config.MinIORootUser + "\"\n" +
        "MINIO_SECRET_KEY=\"" + config.MinIORootPassword + "\"\n" +
        "MINIO_REGION=\"" + config.Region + "\"\n")
    configFile.Close()
    serviceFile, err := os.Create("/etc/systemd/system/minio.service")
    if err != nil {
        fmt.Println("create service file error:", err)
        return false
    }
    serviceFile.WriteString("[Unit]\n" +
        "Description=MinIO\n" +
        "Documentation=https://docs.min.io\n" +
        "Wants=network-online.target\n" +
        "After=network-online.target\n" +
        "\n" +
        "[Service]\n" +
        "User=" + config.SystemUser + "\n" +
        "Group=" + config.SystemUser + "\n" +
        "EnvironmentFile=/etc/default/minio\n" +
        "ExecStart=/usr/local/bin/minio server $MINIO_VOLUMES $MINIO_OPTS\n" +
        "Restart=always\n" +
        "LimitNOFILE=65536\n" +
        "\n" +
        "[Install]\n" +
        "WantedBy=multi-user.target\n")
    serviceFile.Close()

    cmd = exec.Command("systemctl", "daemon-reload")
    err = cmd.Run()
    if err != nil {
        fmt.Println("systemctl daemon-reload error:", err)
        fmt.Println("cmd:", cmd.String())
        return false
    }
    cmd = exec.Command("systemctl", "start", "minio")
    err = cmd.Run()
    if err != nil {
        fmt.Println("systemctl start error:", err)
        fmt.Println("cmd:", cmd.String())
        return false
    }
    cmd = exec.Command("systemctl", "enable", "minio")
    err = cmd.Run()
    if err != nil {
        fmt.Println("systemctl enable error:", err)
        fmt.Println("cmd:", cmd.String())
        return false
    }
    return true
}

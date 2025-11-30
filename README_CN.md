**[简体中文](README_CN.md)** | [English](README.md)

**该 README 绝大部分内容由 [CodeBuddy](https://copilot.tencent.com/) 生成 :)**

---

![GitHub License](https://img.shields.io/github/license/yhlooo/bangbang)
[![GitHub Release](https://img.shields.io/github/v/release/yhlooo/bangbang)](https://github.com/yhlooo/bangbang/releases/latest)
[![release](https://github.com/yhlooo/bangbang/actions/workflows/release.yaml/badge.svg)](https://github.com/yhlooo/bangbang/actions/workflows/release.yaml)

# BangBang

BangBang 是一个去中心的面对面群聊、文件传输工具，通过输入相同 PIN 码即可配对局域网内的多个客户端。

## 特性

- **去中心化架构**：无需中央服务器，所有通信都在对等端之间直接进行
- **PIN 码配对**：只需输入相同的 PIN 码即可连接多个客户端
- **实时聊天**：支持多人参与的面对面群聊
- **文件传输**：在已连接的客户端之间直接分享文件 (TODO)
- **局域网发现**：自动发现同一网络上的其他客户端
- **终端界面**：清晰直观的终端用户界面

## 安装

### Docker

使用镜像 [`ghcr.io/yhlooo/bangbang`](https://github.com/yhlooo/bangbang/pkgs/container/bangbang) 直接 docker run：

```bash
docker run -it --net=host --rm ghcr.io/yhlooo/bang:latest --help
```

### 通过二进制安装

通过 [Releases](https://github.com/yhlooo/bangbang/releases) 页面下载可执行二进制，解压并将其中 `bang` 文件放置到任意 `$PATH` 目录下。

### 从源码编译

要求 Go 1.24.7 或更高版本，执行以下命令下载源码并构建：

```bash
go install github.com/yhlooo/bangbang/cmd/bang@latest
```

构建的二进制默认将在 `${GOPATH}/bin` 目录下，需要确保该目录包含在 `$PATH` 中。

## 使用

### 聊天

要启动或加入聊天会话，使用 `chat` 命令并指定 PIN 码（例如 `7134`）：

```bash
bang chat 7134
```

所有使用相同 PIN 码的参与者都将连接到同一个聊天室。

#### 网络发现

BangBang 使用 UDP 组播地址（默认：`224.0.0.1:7134`）来自动发现同一局域网上的其他客户端。可以使用 `--discovery-addr` 参数自定义发现地址。

#### 日志

运行 `chat` 命令时添加 `--debug` 参数可将日志输出到 `~/.bangbang/bang.log`

### 其他命令

#### 扫描房间

```bash
bang scan
```

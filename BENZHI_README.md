# thermal-cycle-lab__003 Docker 交付说明

## 项目概览
- Thermal Cycle Lab 是一个仅使用 Go 标准库实现的本地 HTTP 服务，用于协调材料试样的热循环实验。实验工程师登记试样并控制运行状态，设备观察员提交传感帧，质量研究员读取阶段报告。所有数据保存在进程内存中，不依赖外部网络、数据库或设备。
- Go module: `thermal-cycle-lab`

## 标准命令

```bash
go build ./...
go test ./...
```

## 实际启动入口

```bash
go run ./cmd/server
```

## Docker 构建

```bash
./build_benzhi_docker.sh thermal-cycle-lab__003-benzhi linux/amd64
docker run --rm -it thermal-cycle-lab__003-benzhi bash
```

## 环境

- 基础镜像: `golang:1.26.2`
- 依赖在镜像构建阶段预下载，容器内可直接执行 Go 构建和测试命令。
- 代码目录: `/app`
- 源码中检测到的服务端口: `8096`, `18096`

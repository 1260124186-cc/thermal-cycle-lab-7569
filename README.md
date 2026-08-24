# Thermal Cycle Lab

Thermal Cycle Lab 是一个仅使用 Go 标准库实现的本地 HTTP 服务，用于协调材料试样的热循环实验。实验工程师登记试样并控制运行状态，设备观察员提交传感帧，质量研究员读取阶段报告。所有数据保存在进程内存中，不依赖外部网络、数据库或设备。

## 功能
- 登记带安全温区的材料试样，并跟踪 ready / in_use / retired 状态。
- 创建多阶段温控曲线，校验目标温度、持续时间和容差。
- 启动、暂停、恢复和完成试验，防止同一试样被重复占用。
- 接收严格递增的传感帧，记录阶段偏差和温度告警。
- 生成按阶段聚合的最终报告，并提供运行面板和事件历史。

## 目录
- `cmd/server`：进程启动、信号处理与 HTTP 生命周期。
- `internal/domain`：试样、曲线、试验、传感帧和报告规则。
- `internal/catalog`：试样与曲线用例。
- `internal/engine`：试验控制、采集、事件和报告生成。
- `internal/store`：并发安全的内存仓储。
- `internal/api`：JSON 路由、编解码、错误映射和中间件。

## 运行
```bash
go run ./cmd/server
```
默认地址为 `:8096`，可通过 `THERMAL_LAB_ADDR=127.0.0.1:18096` 覆盖。健康检查：`curl http://127.0.0.1:8096/health`。

## 构建与测试
```bash
go build ./...
go test ./...
```

## 示例流程
先向 `/api/specimens` 登记试样，再向 `/api/profiles` 创建温控曲线；使用返回的标识调用 `/api/runs` 启动试验，随后向 `/api/runs/{id}/frames` 提交传感帧，最后调用 `/api/runs/{id}/finalize` 并读取 `/api/runs/{id}/report`。

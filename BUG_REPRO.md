# 运行端点绕过校验后 panic复现

## 环境构建与编译

Docker 验证使用当前 `linux/arm64` 平台，镜像内标准编译命令为 `go build ./...`。

## 故障触发命令

```bash
go test -count=1 -run '^TestRunEndpointsRejectUnsafeStateWithoutPanic$' ./internal/api
```

## 修复前实际输出

```text
--- FAIL: TestRunEndpointsRejectUnsafeStateWithoutPanic (0.00s)
    unsafe_run_endpoints_regression_test.go:52: early report status=500 want=404 body={"error":"internal server failure"}
    unsafe_run_endpoints_regression_test.go:56: invalid stage status=202 want=422 body={"data":{"receipt":{"run":{"id":"run-001","specimen_id":"sp-001","profile_id":"pr-001","state":"running","stage_index":7,"accepted_frames":1,"alert_count":0,"started_at":"2026-08-24T08:06:17.667578Z","updated_at":"2026-08-24T08:06:17.667766Z"},"assessment":{"accepted":true,"within_tolerance":true,"delta":0,"message":"frame accepted within stage tolerance"},"cursor":{"sensor_id":"probe-unsafe","last_sequence":1,"last_captured_at":"2026-08-24T08:06:17.667764Z"}},"status":"accepted"}}
    unsafe_run_endpoints_regression_test.go:66: unsafe frame changed accepted_frames=1
    unsafe_run_endpoints_regression_test.go:70: unsafe input escaped validation and panicked during finalize: {"error":"internal server failure"}
FAIL
FAIL	thermal-cycle-lab/internal/api	0.526s
FAIL
```

## 期望行为

缺失报告和越界阶段应按 404/422 拒绝，不能进入 panic 或持久化无效帧。

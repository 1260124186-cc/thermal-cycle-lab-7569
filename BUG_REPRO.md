# 重复传感帧污染运行状态复现

## 环境构建与编译

Docker 验证使用当前 `linux/arm64` 平台，镜像内标准编译命令为 `go build ./...`。

## 故障触发命令

```bash
go test -count=1 -run '^TestDuplicateFrameDoesNotPoisonRunningState$' ./internal/api
```

## 修复前实际输出

```text
2026/08/24 16:06:12 request trace=trace-000001 method=POST path=/api/specimens duration=1ms
2026/08/24 16:06:12 request trace=trace-000002 method=POST path=/api/profiles duration=0s
2026/08/24 16:06:12 request trace=trace-000003 method=POST path=/api/runs duration=0s
2026/08/24 16:06:12 request trace=trace-000004 method=POST path=/api/runs/run-001/frames duration=0s
2026/08/24 16:06:12 request trace=trace-000005 method=POST path=/api/runs/run-001/frames duration=0s
2026/08/24 16:06:12 request trace=trace-000006 method=GET path=/api/runs/run-001 duration=0s
2026/08/24 16:06:12 request trace=trace-000007 method=POST path=/api/runs/run-001/frames duration=0s
duplicate frame rejection changed the public error contract or persisted run state
--- FAIL: TestDuplicateFrameDoesNotPoisonRunningState (0.00s)
    duplicate_frame_regression_test.go:60: duplicate status=500, want 422; body={"error":"advance sensor cursor: sensor frame sequence did not advance"}
    duplicate_frame_regression_test.go:77: run after duplicate: state=paused accepted=1, want running/1
    duplicate_frame_regression_test.go:82: next frame: status=422 body={"error":"apply sensor frame: run does not accept sensor frames in current state"}
FAIL
FAIL	thermal-cycle-lab/internal/api	0.541s
FAIL
```

## 期望行为

重复帧本应返回 422，保持运行状态，并允许后续合法序列继续采集。

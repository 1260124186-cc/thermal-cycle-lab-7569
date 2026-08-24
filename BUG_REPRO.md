# 跨运行完成资源被提前关闭复现

## 环境构建与编译

Docker 验证使用当前 `linux/arm64` 平台，镜像内标准编译命令为 `go build ./...`。

## 故障触发命令

```bash
go test -count=1 -run '^TestSequentialRunsOwnIndependentCompletionResources$' ./internal/api
```

## 修复前实际输出

```text
--- FAIL: TestSequentialRunsOwnIndependentCompletionResources (0.00s)
    sequential_completion_regression_test.go:57: second finalize status=500 body={"error":"internal server failure"}
    sequential_completion_regression_test.go:62: report run-002 status=404 body={"error":"load run report: report for run run-002: record not found"}
FAIL
FAIL	thermal-cycle-lab/internal/api	0.591s
FAIL
```

## 期望行为

两场独立运行都应成功完成，并能分别读取自己的报告。

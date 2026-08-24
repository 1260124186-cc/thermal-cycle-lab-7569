# 已取消采集仍被持久化复现

## 环境构建与编译

Docker 验证使用当前 `linux/arm64` 平台，镜像内标准编译命令为 `go build ./...`。

## 故障触发命令

```bash
go test -count=1 -run '^TestCanceledFrameRequestLeavesNoPersistentEffects$' ./internal/api
```

## 修复前实际输出

```text
2026/08/24 16:06:14 request trace=trace-000001 method=POST path=/api/specimens duration=0s
2026/08/24 16:06:14 request trace=trace-000002 method=POST path=/api/profiles duration=0s
2026/08/24 16:06:14 request trace=trace-000003 method=POST path=/api/runs duration=0s
2026/08/24 16:06:14 request trace=trace-000004 method=POST path=/api/runs/run-001/frames duration=0s
2026/08/24 16:06:14 request trace=trace-000005 method=GET path=/api/runs/run-001 duration=0s
--- FAIL: TestCanceledFrameRequestLeavesNoPersistentEffects (0.00s)
    canceled_frame_regression_test.go:53: canceled frame status=202 want=499 body={"data":{"receipt":{"run":{"id":"run-001","specimen_id":"sp-001","profile_id":"pr-001","state":"running","stage_index":0,"accepted_frames":1,"alert_count":0,"started_at":"2026-08-24T08:06:14.537116Z","updated_at":"2026-08-24T08:06:14.537147Z"},"assessment":{"accepted":true,"within_tolerance":true,"delta":0,"message":"frame accepted within stage tolerance"},"cursor":{"sensor_id":"probe-cancel","last_sequence":1,"last_captured_at":"2026-08-24T08:06:14.537146Z"}},"status":"accepted"}}
    canceled_frame_regression_test.go:64: canceled request persisted run effects: {ID:run-001 SpecimenID:sp-001 ProfileID:pr-001 State:running StageIndex:0 AcceptedFrames:1 AlertCount:0 StartedAt:2026-08-24 08:06:14.537116 +0000 UTC UpdatedAt:2026-08-24 08:06:14.537147 +0000 UTC CompletedAt:<nil> PauseReason:}
    canceled_frame_regression_test.go:75: canceled request persisted frames=1 cursor=1
FAIL
FAIL	thermal-cycle-lab/internal/api	0.468s
FAIL
```

## 期望行为

取消请求本应返回 499，且不写入帧、游标或运行计数。

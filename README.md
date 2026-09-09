# mygo

一个采用常见 Go 项目布局的起始项目。

## 目录说明

- `cmd/`：可执行程序入口；当前入口为 `cmd/mygo`。
- `internal/`：仅供本项目内部使用的代码。
- `pkg/`：可被外部项目复用的公共代码。
- `api/`：API、OpenAPI 或 protobuf 定义。
- `configs/`：配置文件模板及示例。
- `deployments/`：Docker、Kubernetes 等部署配置。
- `scripts/`：构建、开发和自动化脚本。
- `docs/`：项目文档。
- `examples/`：示例代码。
- `test/`：集成测试或端到端测试资源。
- `build/`：构建及打包相关文件。
- `web/`：前端静态资源（如项目需要）。

运行示例：

```bash
go run ./cmd/mygo
```

## 协程示例

`internal/corountine` 是从引擎项目中拷贝并改造成的独立实现，不需要下载外部依赖。它采用协作式调度：协程运行到 `WaitNextFrame`、`Wait`、`WaitForChan` 或 `Yield` 时主动让出执行权，宿主程序通过 `Update` 推进等待任务。

```go
coroutine.Start(nil)
co := coroutine.New(nil)
co.OnInited()

done := make(chan struct{})
co.Create("worker", func(me coroutine.Thread) int {
	// 做一小段工作，然后等待下一帧继续。
	co.WaitNextFrame()
	close(done)
	return 0
})

for {
	co.Update()
	coroutine.Update(1.0/30.0, coroutine.DefaultFPS)
	select {
	case <-done:
		co.AbortAllAndWait(time.Second)
		return
	default:
	}
}
```

其中 `coroutine.Start` 和 `coroutine.Update` 是为脱离原引擎补充的简易时钟；在实际游戏或事件循环中，可以将它们替换为宿主自己的帧计时器。

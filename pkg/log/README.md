## 日志使用说明
> 本库基于[zap](go.uber.org/zap)库二次封装而成，使用[lumberjack](github.com/natefinch/lumberjack)做的日志分割。
### 例子
> 例子[详情](./example_test.go)

```
func main() {
	log.Init(&log.Config{
		Level:      log.LevelInfo,
		Filename:   "test.log",
		MaxSize:    10,
		MaxAge:     1,

		Compress:   false,
		DevMode:    false,
	})

	people := "Alice"
	log.Debug("Hello", "people", people)
	log.Info("Hello", "people", people)
	log.Warn("Hello", "people", people)
	log.Error("Hello", "people", people)
}
```

### 其他特性
- 支持日志彩色输出
> 上述例子中的```log.Config.DevMode```若为```true```则会开启彩色输出,
>为```false```则会关闭彩色输出，在生产环境需配置为```false```。

 ![auater](./colourful_log.png)

- 默认配置
> 若不对日志进行任何配置，直接 ```import``` 此库，则会启用默认配置，默认配置时不会启用动态调整日志等级的功能。
> 默认配置详情如下：

```
Level:      LevelInfo,
Filename:   "./log/default.log",
MaxSize:    100, // MB
MaxAge:     30,  // days
MaxBackups: 10,  // files
Compress:   false,
DevMode:    false,
```

### 其他说明
> 若```log.Config.DevMode == true```则会将日志同时输出到日志文件和控制台，且日志格式为有利于阅读的格式。<br>
> 若```log.Config.DevMode == false```则只会在将日志输出到文件，不会在控制台展示，且日志格式为JSON。<br>
> ```log.Config.DevMode```配置项与日志等级无关，此配置项仅作为方便开发人员调试使用。

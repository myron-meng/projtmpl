/*

Package log 是标准库的替代，使用库 [phuslu/log](https://github.com/phuslu/log), 此库优点有：
- 可以输出日志到控制台或者文件，控制台日志支持颜色高亮区分不同级别的日志
- 支持以 JSON 格式输出日志
- 支持日志文件达到指定的文件大小后自动轮转
- 支持指定日志的清理策略
- 零依赖且功能完备
- 配置和使用方式非常简单
- 高性能，比 [uber/zap](https://github.com/uber-go/zap) 还高

*/
package log

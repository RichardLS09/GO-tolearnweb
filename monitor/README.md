Monitor
====

监控模块可以通过开启的指定端口监控系统状态。

可以被监控的模块都实现了```monitor.Observable```接口。
注册这些模块就可以通过监控接口获得运行时状态数据。
```
monitor.Observe("queue", queue.GetQueueContainer())
monitor.Observe("crontab", crontab.GetCrontabContainer())
monitor.Observe("producer", queue.GetProducerContainer())
monitor.Observe("consumer", queue.GetConsumerContainer())
go monitor.Run("127.0.0.1:9998")
```

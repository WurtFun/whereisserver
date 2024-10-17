# Whereisserver

- forked form [Tnze/go-mc](https://github.com/Tnze/go-mc)
- 需要环境 go 1.22

### 如何使用?
- 请指定起始IP段和结束IP段进行遍历

 例如：`go run main.go <起始IPv4地址> <结束IPv4地址>`

# 日志内容
所有返回内容将被保存在`catch-server-list/laster-catch.txt`文件中。每次运行时，旧的日志文件会自动重命名为`catched_<timestamp>.txt`形式的新文件。

# 许可
本项目遵循[MIT License](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/licensing-a-repository#disclaimer)。

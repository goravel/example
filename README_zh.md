<p align="center"><img src="https://www.goravel.dev/logo.png" width="300"></p>

[English](./README.md) | 中文

# 关于 Goravel

Goravel 是一个功能完备、具有良好扩展能力的 Web 应用程序框架。作为一个起始脚手架帮助 Golang 开发者快速构建自己的应用。

框架风格与 [Laravel](https://github.com/laravel/laravel) 保持一致，让 PHPer 不用学习新的框架，也可以愉快的玩转 Golang！致敬
Laravel！

欢迎 Star, PR, Issues！

## 快速上手

### 启动服务

`go run .` or `air`

[关于 air]: https://www.goravel.dev/zh/getting-started/installation.html#热更新

### DB

[app/http/controllers/db_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/db_controller.go)

### Websocket

[app/http/controllers/websocket_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/websocket_controller.go)

关于分布式 Websocket 可参考文章：https://learnku.com/articles/39701

### Validation

[app/http/controllers/validation_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/validation_controller.go)

### JWT Middleware

[app/http/controllers/jwt_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/jwt_controller.go)

### Unit Test (Testing With Mock)

[app/http/controllers/validation_controller_test.go](https://github.com/goravel/example/blob/master/app/http/controllers/validation_controller_test.go)

### Integration Test (Testing With Configuration)

[tests/controllers/validation_controller_test.go](https://github.com/goravel/example/blob/master/tests/controllers/validation_controller_test.go)

### GRPC

[app/grpc/controllers/user_controller.go](https://github.com/goravel/example/blob/master/app/grpc/controllers/user_controller.go)

### Swagger

[app/http/controllers/swagger_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/swagger_controller.go)

### 单页面前端应用集成到框架

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L43)

### 视图嵌套

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L52)

### 本地化

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L60)

### Session

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L65)

### Cookie

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L81)

## 文档

在线文档 [https://www.goravel.dev/zh](https://www.goravel.dev/zh)

> 优化文档，请提交 PR 至文档仓库 [https://github.com/goravel/docs](https://github.com/goravel/docs)

## 群组

微信入群，请备注 Goravel

<p align="left"><img src="https://www.goravel.dev/wechat.jpg" width="200"></p>

## 开源许可

Goravel 框架是在 [MIT 许可](https://opensource.org/licenses/MIT) 下的开源软件。

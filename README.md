<p align="center"><img src="https://www.goravel.dev/logo.png?v=1.14.x" width="300"></p>

English | [中文](./README_zh.md)

## About Goravel

Goravel is a web application framework with complete functions and good scalability. As a starting scaffolding to help Gopher quickly build their own applications.

The framework style is consistent with [Laravel](https://github.com/laravel/laravel), let Phper don't need to learn a new framework, but also happy to play around Golang! Tribute Laravel!

Welcome to star, PR and issues！

## Getting Started

### Start Service

`go run .` or `air`

[About air]: https://www.goravel.dev/getting-started/installation.html#live-reload

### DB

[app/http/controllers/db_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/db_controller.go)

### Websocket

[app/http/controllers/websocket_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/websocket_controller.go)

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

### Integration of single page application into the framework

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L41)

### View nesting

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L52)

### Localization

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L60)

### Session

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L65)

### Cookie

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L81)

## Documentation

Online documentation [https://www.goravel.dev](https://www.goravel.dev)

> To optimize the documentation, please submit a PR to the documentation
> repository [https://github.com/goravel/docs](https://github.com/goravel/docs)

## Group

Welcome more discussion in Telegram.

[https://t.me/goravel](https://t.me/goravel)

<p align="left"><img src="https://www.goravel.dev/telegram.jpg" width="200"></p>

## License

The Goravel framework is open-sourced software licensed under the [MIT license](https://opensource.org/licenses/MIT).

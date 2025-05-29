## 介绍

使用 Gin 框架实现基本的用户管理和基于 JWT 的 Token 鉴权功能, 部分代码使用 sponge 生成.

数据库默认使用 Sqlite, 缓存默认使用内存数据库, 可以通过配置文件切换为 MySQL 和 Redis.

生成Swagger文档: `make docs`  
运行命令: `make run`

## Swagger API文档

运行程序后访问以下地址查看文档:

地址: http://127.0.0.1:8080/swagger/index.html

修改代码后需要重新生成文档, 生成命令: `make docs`

## TODO

功能:

- 增加基于casbin权限认证逻辑

已知问题:

自动续期: 如果客户端在旧 Token 即将过期时同时发出多个请求，这些请求可能都会触发“重新生成 JWT”的逻辑。这可能导致：

- 生成多个不同的新 JWT。
- 缓存被多次更新，可能存在竞争条件。
- 客户端可能在短时间内收到多个包含新 Token 的响应，需要客户端逻辑来处理接收到的最新有效 Token。

sqlite驱动必须支持cgo, 修改纯go实现sqlite

types中, 自定义单独的id类型, 给其它类型使用, 以避免雪花ID的前端溢出问题
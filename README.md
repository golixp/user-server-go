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


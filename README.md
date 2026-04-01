
* Для генерации файлов устанавливаем [Свагер](https://github.com/go-swagger/go-swagger/releases)

<img width="1129" height="584" alt="image" src="https://github.com/user-attachments/assets/b4342a70-2df7-41f5-aaa0-6e82df4dbac7" />

Структура проекта на 18.02:

├── api
│   ├── helpers.go
│   ├── hndlrs_others.go
│   ├── manager.go
│   ├── stub.go
│   └── wrappers.go
├── cmd
│   └── main
│       └── main.go
├── entities
│   ├── languages.go
│   └── notify.go
├── example.env
├── go.mod
├── go.sum
├── internal
│   ├── models
│   │   ├── config
│   │   ├── user.go
│   │   └── verification_code.go
│   ├── repository
│   │   ├── mysql_user_repository.go
│   │   └── user_repository.go
│   ├── utils
│   │   ├── code_generator.go
│   │   ├── password.go
│   │   └── validator.go
│   └── web
│       └── register.html
├── main.go
├── Makefile
├── manager
│   └── manager.go
├── migrations
│   ├── 001_create_users.sql
│   ├── 002_create_verification_codes.sql
│   ├── go.mod
│   └── run.go
├── mysql
│   ├── client.go
│   └── stub.go
├── mysql_user_repository.go
├── README.md
├── swagger.yml
├── user.go
├── user_repository.go
└── web
    └── templates
        └── register.html

15 directories, 35 files

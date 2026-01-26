
* Для генерации файлов устанавливаем [Свагер](https://github.com/go-swagger/go-swagger/releases)

<img width="1129" height="584" alt="image" src="https://github.com/user-attachments/assets/b4342a70-2df7-41f5-aaa0-6e82df4dbac7" />


1) HTTP сервер (main.go) — принимает запросы
2) Репозитроий MySQL — хранит пользователей и токены
3) Auth сервис — бизнес-логика аутентификации
4) Email sender — отправляет magic-link на почту
5) WebSocket hub — для real-time уведомлений
6) HTML шаблоны + JavaScript = взаимодействие с пользователем

Структура проекта на 26.01:
coursework/
├── cmd/
│       └── main.go
│
├── internal/
│   ├── handlers/            ← HTTP (GET / POST)
│   │   └── user_handler.go
│   │
│   │
│   ├── repository/          ← работа с БД
│   │   ├── user_repository.go       
│   │   └── mysql_user_repository.go 
│   │
│   └── models/
│       └── user.go
│
├── web/
│   ├── templates/           ← тут HTML
│   │   └── register.html
│   │
│   └── static/              ← позже будет CSS / JS
│       └── style.css
│
├── .env
├── .gitignore
├── go.mod
├── go.sum
└── README.md

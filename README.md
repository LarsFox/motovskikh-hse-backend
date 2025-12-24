
* Для генерации файлов устанавливаем [Свагер](https://github.com/go-swagger/go-swagger/releases)

Общая архитектура ЛК разработчика тестов:
Пользователь → [Frontend: JS] → [Backend: Go] → [БД: MySQL] → [Email: SMTP]
                    ↑                              ↑
                    └──── [WebSocket: Real-time] ←─┘

HTTP Request → [Server] → [Router] → [Handler] → [AuthService]
                                                       ↓
                [MySQLRepository] ← Данные пользователя
                       ↓
                [TokenService] ← Генерация токенов
                       ↓
                [EmailSender] ← Отправка письма
                       ↓
                [WebSocketHub] ← Уведомление о входе

1) HTTP сервер (main.go) — принимает запросы
2) Репозитроий MySQL — хранит пользователей и токены
3) Auth сервис — бизнес-логика аутентификации
4) Email sender — отправляет magic-link на почту
5) WebSocket hub — для real-time уведомлений
6) HTML шаблоны + JavaScript = взаимодействие с пользователем

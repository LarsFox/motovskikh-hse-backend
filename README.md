* Для генерации файлов устанавливаем [Свагер](https://github.com/go-swagger/go-swagger/releases)

├── api/                    	# HTTP handlers (слой представления)
│   ├── helpers.go          	# Вспомогательные функции
│   ├── hndlrs_others.go    	# Обработчики ошибок
│   ├── hndlrs_submit.go    	# Обработчики статистики
│   ├── manager.go          	# Менеджер API (роутинг)
│   └── wrappers.go         	# Middleware
├── cmd/main/              		# Точка входа
│   └── main.go           		# Инициализация приложения
├── entities/              		# Модели данных
│   ├── languages.go        	# Константы языков
│   └── notify.go           	# Логирование ошибок
│   └── analysis.go
│   └── percent_bucket.go
│   └── time_bucket.go
│   └── test_stats.go
├── generated/            		# Сгенерированный код Swagger
│   └── models/          		  # Модели из swagger.yml
│   └── mocks/
├── manager/              		# Бизнес-логика (service layer)
│   └── manager.go       		  # Основная логика
│   └── calculations.go          # Вычисления
│   └── submit.go       		  # Основной метод
│   └── validation.go       	  # Валидация
│   └── updates.go       		  # Обновления БД
├── mysql/               		  # Работа с БД (data access layer)
│   ├── client.go       		  # Клиент MySQL
│   └── client_bucket.go 		  # Работа с бакетами и БД
│   └── db_stats.go 		      # Работа со статистикой в БД
└── doc.json            		  # Swagger документация

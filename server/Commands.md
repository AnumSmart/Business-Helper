команда для windows для установки goose:

1. для win7, находясь в корне проекта выполняем коменду: set GOBIN=%cd%\bin&&go install github.com/pressly/goose/v3/cmd/goose@м3.20.0
   для win10: set GOBIN=%cd%\bin&&go install github.com/pressly/goose/v3/cmd/goose@latest
   не забываем добавить папку /bin в .gitignore
2. после установки бинарника goose нужно создать папку /migrations в корне проекта
3. нужно создать файл миграции. В корне проекта выполняем команду: .\bin\goose.exe -dir .\migrations create create_messages_table sql
4. заполняем файл миграции
5. Применить миграции (up): .\bin\goose.exe -dir .\migrations postgres "user=postgres password=ваш*пароль dbname=ваша*база sslmode=disable" up
6. Откатить последнюю миграцию (down): .\bin\goose.exe -dir .\migrations postgres "user=postgres password=ваш*пароль dbname=ваша*база sslmode=disable" down
7. Проверить статус миграций: .\bin\goose.exe -dir .\migrations postgres "user=postgres password=ваш*пароль dbname=ваша*база sslmode=disable" status

# Отчет по Ленте новостей социальной сети

## Содержание

1. [ Задача ](#task)
2. [ Архитектура проекта ](#architecture)
    - [ Сервис Событий / Новостей ](#news-service)
    - [ Кэш сервиса событий ](#cache)
        - [ Разогрев кэша ](#cache-heating)
        - [ Инвалидация кэша ](#cache-invalidation)
        - [ Структуры кэша ](#cache-structures)
    - [ Процессы-обработчики событий ](#processors)
3. [ Запуск ](#startup)
4. [ Тестирование ](#tests)
6. [ Выводы ](#total)

<a name="task"></a>

## Задача

Разработать ленту новостей.

- Создается отдельная страница, куда пишутся все обновления друзей;
- Лента формируется на уровне кешей;
- Формирование ленты производится через постановку задачи в очередь на часть подписчиков;
- В ленте держать последние 1000 обновлений друзей;
- Лента должна кешироваться.

Требования:

- Верно работает инвалидация кеша.
- Обновление лент работает через очередь;
- Есть возможность перестройки кешей из СУБД.

<a name="architecture"></a>

## Архитектура проекта

![Архитектура сервисов](project-arch.png)

- **REST API Server** - доступная клиентам точка входа с RESTful API интерфейсом на борту;
- **MySQL** - сервер Базы Данных. Используется REST API сервером, для хранения различных сущностей и их состояний;
- **Kaffka** - Очередь сообщений в лице Kafka, выполняет роль шины событий;
- **Events Service (News Service)** - Сервис для хранения/чтения пользовательских событий;
- **Redis Cache** - кэш, используется Events Service, для быстрой выдачи списка новостей(событий) и быстрого получения
  списка подписчиков на записи события;
- **MongoDB** - NoSQL база данных, используется Events Service для персистентного хранения событий и подписчиков.

Клиенты шлют запросы к REST API.

Если запрос генерирует одно из пользовательских событий, то событие оформляется в сообщение и улетает в очередь (Kafka).

Из очереди сообщение достает сервис новостей и осуществляет конечную обработку события - пишет в персистентное
хранилищие и обновляет кэш.

<a name="news-service"></a>

### Сервис Событий / Новостей

Сервис событий держит на борту данные о связях пользователей с их подписчиками + списки событий, слинкованных с
подписчиками **на момент создания** события.

При старте сервиса:

- разогревается кэш со списками подписчиков;
- взлетает заданное в конфиге число процессов-потребителей очереди (эквивалентно числу партиций в топике)
- взлетает воркеров-обработчиков событий числом равным количеству процессов-потребителей, умноженному на количество
  типов событий.

Процесс потребитель получает сообщение с событием, проводит первичную валидацию и конвертацию события в модель, после
чего моделька улетает во внутренний канал событий, откуда ее подхватывает менеджер процессов-обработчиков и проксирует
событие в канал подходящего обработчика.

Процессы обработчики сохраняют события в персистентное хранилище и поддерживают актуальное состояние кэша. События не
имеющие связей с подписчиками (у пользователя нет подписчиков), отбрасываются.

На стороне сервиса открыт HTTP интерфейс, к которому обращается REST API Server (Backend Service), для чтения событий
связанных с пользователем.

При чтении списка событий, события отдаются из кэша, или читаются напрямую из БД, если кэш отсутствует.

**Код сервиса новостей можно найти по ссылке https://github.com/yelsukov/otus-ha/tree/main/news**

<a name="cache"></a>

### Кэш для сервиса событий

<a name="cache-heating"></a>

#### Разогрев кэша

Кэш со списком подписчиков разогревается при старте сервиса.

Кэш со списком событий разогревается индивидуально для каждого пользователя, при регистрации события входа (login).

<a name="cache-invalidation"></a>

#### Инвалидация кэша

Кэш подписчиков всегда синхронизирован с персистентым хранилищем. Инваидация кэша подписчиков происходит только при
старте сервиса, перед разогревом.

Инвалидация кэша событий происходит по событию выхода пользователя из системы (logout).

<a name="cache-structures"></a>

#### Структуры кэша

В качестве структуры взят список, позволяющий практически для всех необходимых операций поддерживать сложность O(1).

Так же списки позволяют легко ограничивать количество содержащихся в них элементов.

В данной реализации присутствуют два вида списков:

1. Список новостей пользователя. Ограничен одной тысячей записей;
2. Список подписчиков пользователя. Не ограничен.

<a name="processors"></a>

### Процессы-обработчики событий

В данный момент существует четыре вида обработчиков событий login, logout, follow и unfollow.

Все обработчики, помимо определенного для них набора действий, выполняют операции сохранения события в персистентном
хранилище + раскидывают событие по спискам подписчиков в кэше.

Если пользователь создавший событие не имеет подписчиков, то событие игнорируется.

**Обработчик login события**

Выполняет разогрев кэша со списком событий для вошедшего в систему пользователя.

**Обработчик logout события**

Выполняет инвалидацию кэша со списком событий для вышедшего из системы пользователя.

**Обработчики follow/unfollow событий**

Выполняют обновление списка подписчиков в персистентном хранилище и в кэше.

<a name="startup"></a>

## Запуск

Проект стартуется командой:

> make upFull

или через docker-compose:
> sudo docker-compose -f docker-compose.standalone.yml -f docker-compose.news.yml -f docker-compose.infra.yml up --build -d

После запуска проверяем логи сервиса Новостей:

> sudo docker logs news

Должны увидеть там следующее:

```
time="2021-02-22 07:34:14" level=info msg="found member: news@a8436c2937e1 (github.com/segmentio/kafka-go)-4d0daa17-0e7b-4b8c-be6f-15d328d9d258/[]byte(nil)"
time="2021-02-22 07:34:14" level=info msg="found member: news@a8436c2937e1 (github.com/segmentio/kafka-go)-79a2c30d-626e-4bca-a77c-c9d9ed28163c/[]byte(nil)"
time="2021-02-22 07:34:14" level=info msg="found member: news@a8436c2937e1 (github.com/segmentio/kafka-go)-fbe44364-3278-4fee-8266-bc0b1ee42711/[]byte(nil)"
time="2021-02-22 07:34:14" level=info msg="found topic/partition: userEvents/0"
time="2021-02-22 07:34:14" level=info msg="found topic/partition: userEvents/2"
time="2021-02-22 07:34:14" level=info msg="found topic/partition: userEvents/1"
time="2021-02-22 07:34:14" level=info msg="assigned member/topic/partitions news@a8436c2937e1 (github.com/segmentio/kafka-go)-4d0daa17-0e7b-4b8c-be6f-15d328d9d258/userEvents/[0]"
time="2021-02-22 07:34:14" level=info msg="assigned member/topic/partitions news@a8436c2937e1 (github.com/segmentio/kafka-go)-79a2c30d-626e-4bca-a77c-c9d9ed28163c/userEvents/[2]"
time="2021-02-22 07:34:14" level=info msg="assigned member/topic/partitions news@a8436c2937e1 (github.com/segmentio/kafka-go)-fbe44364-3278-4fee-8266-bc0b1ee42711/userEvents/[1]"
...
time="2021-02-22 07:34:14" level=info msg="initializing kafka reader for partition 2 of userEvents starting at offset -2"
time="2021-02-22 07:34:14" level=info msg="started heartbeat for group, event.listeners [3s]"
time="2021-02-22 07:34:14" level=info msg="started commit for group event.listeners\n"
time="2021-02-22 07:34:14" level=info msg="initializing kafka reader for partition 1 of userEvents starting at offset -2"
time="2021-02-22 07:34:14" level=info msg="started heartbeat for group, event.listeners [3s]"
time="2021-02-22 07:34:14" level=info msg="started commit for group event.listeners\n"
time="2021-02-22 07:34:14" level=info msg="started commit for group event.listeners\n"
time="2021-02-22 07:34:14" level=info msg="initializing kafka reader for partition 0 of userEvents starting at offset -2"
time="2021-02-22 07:34:14" level=info msg="started heartbeat for group, event.listeners [3s]"
time="2021-02-22 07:34:14" level=info msg="the kafka reader for partition 2 of userEvents is seeking to offset 0"
time="2021-02-22 07:34:14" level=info msg="the kafka reader for partition 1 of userEvents is seeking to offset 0"
time="2021-02-22 07:34:14" level=info msg="the kafka reader for partition 0 of userEvents is seeking to offset 0"
```

Это означает, что процессы-потребители для очереди событий успешно взлетели, подключились к брокеру, и каждый получил
свою партицию для чтения сообщений.

<a name="tests"></a>

## Тестирование

1) **Регистрируем пользователя через GUI**

Открываем фронтенд, по адресу http://127.0.0.1:8008/registration и регистрируемся:

![Регистрация](registration.png)

Выполняем вход под пользовтаелем `tester1`, с паролем `1234567890`

2) **Прямыми запросами к REST API регистрируем еще двух пользователей**

> curl -XPOST http://127.0.0.1:8007/v1/auth/sign-up -d '{"username":"tester2", "password":"1234567890", "passwordConfirm":"1234567890", "gender":"m"}'

и

> curl -XPOST http://127.0.0.1:8007/v1/auth/sign-up -d '{"username":"tester3", "password":"1234567890", "passwordConfirm":"1234567890", "gender":"m"}'

3) **Через GUI подписываемся пользоватем `tester1` на обновления пользователей `tester2` и `tester3`**

![Подписались на tester2](follow-tester2.png) ![Подписались на tester3](follow-tester3.png)

4) **Авторизируемся под пользователями `tester2` и `tester3`**

Выполняем прямые запросы к REST API

> curl -XPOST http://127.0.0.1:8007/v1/auth/sign-in -d '{"username":"tester2", "password":"1234567890"}'

```json
{
  "object": "token",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJleHAiOjE2MTM5OTg0NDMsImlhdCI6MTYxMzk5NDg0MywiaXNzIjoiaGEtYmFja2VuZCJ9.ucmuJVEgFwuhWhIiYVHy1SShwRxVQC1KVXS2w7ORoqY",
  "userId": 4,
  "username": "tester2",
  "expires_at": 1613998443
}
```

> curl -XPOST http://127.0.0.1:8007/v1/auth/sign-in -d '{"username":"tester3", "password":"1234567890"}'

```json
{
  "object": "token",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo1LCJleHAiOjE2MTM5OTg0NDcsImlhdCI6MTYxMzk5NDg0NywiaXNzIjoiaGEtYmFja2VuZCJ9.tGqDjA0WCfe16VqqlNRB8mGUCiy-E6EpHI3SmJkSSCc",
  "userId": 5,
  "username": "tester3",
  "expires_at": 1613998447
}
```

5) **Пользователем `tester2` подписываемся на обновления пользовтелей `tester1` `tester3`**

> curl -XPOST http://127.0.0.1:8007/v1/friends/3  -H "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJleHAiOjE2MTM5OTg0NDMsImlhdCI6MTYxMzk5NDg0MywiaXNzIjoiaGEtYmFja2VuZCJ9.ucmuJVEgFwuhWhIiYVHy1SShwRxVQC1KVXS2w7ORoqY"

> curl -XPOST http://127.0.0.1:8007/v1/friends/5  -H "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJleHAiOjE2MTM5OTg0NDMsImlhdCI6MTYxMzk5NDg0MywiaXNzIjoiaGEtYmFja2VuZCJ9.ucmuJVEgFwuhWhIiYVHy1SShwRxVQC1KVXS2w7ORoqY"

6) **Пользователем `tester3` подписываемся на обновления пользователя `tester2`**

> curl -XPOST http://127.0.0.1:8007/v1/friends/4  -H "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo1LCJleHAiOjE2MTM5OTg0NDcsImlhdCI6MTYxMzk5NDg0NywiaXNzIjoiaGEtYmFja2VuZCJ9.tGqDjA0WCfe16VqqlNRB8mGUCiy-E6EpHI3SmJkSSCc"

7) **Проверяем кэш сервиса новостей**

Подключаемся к redis-клиенту в docker-контейнере:
> sudo docker exec -it redis redis-cli

Выводим список ключей:
> 127.0.0.1:6379> keys *

```
1) "user:3:followers"
2) "user:5:followers"
3) "user:4:events"
4) "user:4:followers"
5) "user:3:events"
```

Видим в наличии кэш подписчиков для трех активных пользователей и кэш событий для двух пользователей, имеющих
подписчиков.

> Для пользователя `tester3` новости отсутствуют, т.к. он подписался уже после того, как `tester2` испустил события login и follow.

Прочтем подписчиков `tester2`:
> 127.0.0.1:6379> lrange user:4:followers 0 -1

```
1) "5"
2) "3"
```

Видим, что на события подписаны `tester3` и `tester1`.

Прочтём события по подписчику `tester1`:
> 127.0.0.1:6379> lrange user:3:events 0 -1

```
1) "{\"name\":\"follow\",\"message\":\"tester3 now follows for the tester2\",\"uid\":5}"
2) "{\"name\":\"follow\",\"message\":\"tester2 now follows for the tester3\",\"uid\":4}"
3) "{\"name\":\"follow\",\"message\":\"tester2 now follows for the tester1\",\"uid\":4}"
4) "{\"name\":\"login\",\"message\":\"tester3 just logged in\",\"uid\":5}"
5) "{\"name\":\"login\",\"message\":\"tester2 just logged in\",\"uid\":4}"
```

8) **Производим logout для пользователей `tester2` и `tester3`**

> curl -XPOST http://127.0.0.1:8007/v1/auth/sign-out -H "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJleHAiOjE2MTM5OTg0NDMsImlhdCI6MTYxMzk5NDg0MywiaXNzIjoiaGEtYmFja2VuZCJ9.ucmuJVEgFwuhWhIiYVHy1SShwRxVQC1KVXS2w7ORoqY"

> curl -XPOST http://127.0.0.1:8007/v1/auth/sign-out -H "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo1LCJleHAiOjE2MTM5OTg0NDcsImlhdCI6MTYxMzk5NDg0NywiaXNzIjoiaGEtYmFja2VuZCJ9.tGqDjA0WCfe16VqqlNRB8mGUCiy-E6EpHI3SmJkSSCc"

Кэш событий для этих пользователей должен быть инвалидирован. Проверим это.

Прочтем логи сервиса новостей
> sudo docker logs new

Обнаружим там следуюшие записи:

```
...
time="2021-02-22 11:59:07" level=debug msg="storing `logout` event into db"
time="2021-02-22 11:59:07" level=debug msg="got 2 followers from Cache for user #4"
time="2021-02-22 11:59:07" level=debug msg="adding `logout` event to all followers of user #4"
time="2021-02-22 11:59:07" level=debug msg="Invalidating cache on logout"
...
time="2021-02-22 11:59:40" level=debug msg="storing `logout` event into db"
time="2021-02-22 11:59:40" level=debug msg="got 2 followers from Cache for user #5"
time="2021-02-22 11:59:40" level=debug msg="adding `logout` event to all followers of user #5"
time="2021-02-22 11:59:40" level=debug msg="Invalidating cache on logout"
...
```

Из лога видно, что при записи события в БД, список подписчиков был взят из кэша. Далее событие было распределено по всем
закэшированным спискам у подписчиков автора данного события. В конце процедуры, кэш был инвалидирован для автора
события, т.к. была совершена процедура выхода.

Соединимся с redis клиентом и проверим напрямую в кэше:
> 127.0.0.1:6379> keys *

```
1) "user:3:followers"
2) "user:5:followers"
3) "user:4:followers"
4) "user:3:events"
```

Как видим, кэш новостей для пользователей `tester2` и `tester3`, был действительно удален.

9) **Запрашиваем новости для `tester1`**

> curl http://127.0.0.1:8007/v1/news -H "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJleHAiOjE2MTM5OTgzNjcsImlhdCI6MTYxMzk5NDc2NywiaXNzIjoiaGEtYmFja2VuZCJ9.OGBWvw0BR0W_W2jmf_qME-p0rFy9PhjXMm0gExaLo_M"

```json
{
  "object": "list",
  "data": [
    {
      "object": "event",
      "name": "logout",
      "message": "tester3 just logged out",
      "ts": 1613999060
    },
    {
      "object": "event",
      "name": "logout",
      "message": "tester2 just logged out",
      "ts": 1613999041
    },
    {
      "object": "event",
      "name": "follow",
      "message": "tester3 now follows for the tester2",
      "ts": 1613998989
    },
    {
      "object": "event",
      "name": "follow",
      "message": "tester2 now follows for the tester1",
      "ts": 1613998951
    },
    {
      "object": "event",
      "name": "follow",
      "message": "tester2 now follows for the tester3",
      "ts": 1613998774
    },
    {
      "object": "event",
      "name": "login",
      "message": "tester3 just logged in",
      "ts": 1613998713
    },
    {
      "object": "event",
      "name": "login",
      "message": "tester2 just logged in",
      "ts": 1613998709
    }
  ]
}
```

Этот же список можно увидеть в GUI, на странице Dashboard:
![Список новостей](final-dashboard.png)

10) **Производим logout для `tester1`**

В redis консоли проверяем, что кэш новостей так же был инвалидирован:
> 127.0.0.1:6379> keys *

```
1) "user:3:followers"
2) "user:5:followers"
3) "user:4:followers"
```

11) **Производим повторный вход для `tester1`**.

**Проверяем, что кэш греется.**

Читаем логи сервиса новостей

```
time="2021-02-22 12:06:42" level=debug msg="storing `login` event into db"
...
time="2021-02-22 12:06:42" level=debug msg="Heating events for user#3"
time="2021-02-22 12:06:43" level=debug msg="got 7 events from DB for user #3"
time="2021-02-22 12:06:43" level=debug msg="7 events where been added to cache"
```

Проверим в redis консоли:
> 127.0.0.1:6379> keys *

```
1) "user:3:followers"
2) "user:5:followers"
3) "user:4:followers"
4) "user:3:events"
```

> 127.0.0.1:6379> lrange user:3:events 0 1000

```
1) "{\"name\":\"logout\",\"message\":\"tester3 just logged out\",\"uid\":5}"
2) "{\"name\":\"logout\",\"message\":\"tester2 just logged out\",\"uid\":4}"
3) "{\"name\":\"follow\",\"message\":\"tester3 now follows for the tester2\",\"uid\":5}"
4) "{\"name\":\"follow\",\"message\":\"tester2 now follows for the tester3\",\"uid\":4}"
5) "{\"name\":\"follow\",\"message\":\"tester2 now follows for the tester1\",\"uid\":4}"
6) "{\"name\":\"login\",\"message\":\"tester3 just logged in\",\"uid\":5}"
7) "{\"name\":\"login\",\"message\":\"tester2 just logged in\",\"uid\":4}"
```

Видим, что кэш действительно разогрелся.

**Проверяем, что события читаются из разогретого кэша**

Для этого запрашиваем события через API, или заходим на страницу Dashboard.

Читаем логи сервиса новостей:

```
time="2021-02-22 12:09:32" level=debug msg="got 7 events from Cache for user #3"
2021/02/22 12:09:32 [c4cda7502da4/npJOG1wOdn-000002] "GET http://news:8082/events?fid=3 HTTP/1.1" from 172.30.0.8:41098 - 200 182B in 386.548µs
```

Видим, что данные взяты из кэша, а обработка запроса от его получения до ответа заняла 386.548 наносекунд.

<a name="total"></a>

## Выводы

- Kafka хорший инструмент, если ею правильно забивать гвозди.
- Redis хороший инструмент, если им правильно забивать гвозди.
- В данной реализации имеется несколько огрехов, в том числе с избыточностью и синхронизацеий данных. Решил, что для
  non-production версии можно принять эти условности за норму.
# Отчет по домашнему заданию "In-Memory СУБД"

## Содержание

1. [ Задача ](#task)
2. [ Подготовка окружения ](#env-prepearings)
3. [ Описание сервисов и настроек ](#configuration)
   - [ Replicator ](#conf-replicator)
   - [ Tarantool ](#conf-tarantool)
   - [ Mysql ](#conf-mysql)
4. [ Нагрузочное тестирование ](#load-tests)
   - [ Что тестируем ](#tests-what)
   - [ Чем и как тестируем ](#tests-how)
   - [ Подготовка ](#tests-prepearings)
   - [ Прогон тестов ](#testing)
      - [ Mysql ](#tests-mysql)
      - [ Tarantool ](#tests-tarantool)
   - [ Сводные таблицы ](#pivot-tables)
   - [ Мысли по результатам ](#tests-results)
5. [ Выводы ](#total)

<a name="task"></a>

### Задача

Репликация из MySQL в tarantool:

1) Выбрать любую таблицу, которую мы читаем с реплик MySQL.
2) С помощью программы https://github.com/tarantool/mysql-tarantool-replication настроить реплицирование в tarantool (
   лучше всего версии 1.10).
3) Выбрать любой запрос и переписать его на lua-процедуру на tarantool.
4) Провести нагрузочное тестирование, сравнить tarantool и MySQL по производительности.

Требования:

- Репликация из MySQL в tarantool работает.
- Хранимые процедуры в tarantool написаны корректно.
- Хранимые процедуры выполнены по code style на примере репозитория Mail.Ru.
- Нагрузочное тестирование проведено.

<a name="env-prepearings"></a>

## Подготовка окружения

1. Стартуем окружение командой

> make startTaran

Или через
> sudo docker-compose -f docker-compose.tarantool.yml -f docker-compose.infra.yml up --build -d

Наполняем мускул данными
> make seed SEED_QTY=1000000

На этом все, окружение готово к тестам.

Для backend API не успел сделать нормальное переключение между MySQL и Tarantool. Произвести замену можно собрав сервис
без переменной окружения `TARAN_DSN`.

<a name="configuration"></a>

## Описание сервисов и настроек

<a name="conf-replicator"></a>

### Replicator

Тулзой предложенной в задании пользоваться не стал.

Собрал ее в контейнере без проблем, но дальше, как бы я не старался, она у меня так и не завелась, ни в качестве
системного сервиса, ни в качестве отдельно стартующего бинарника.

Расчехлил гугль, поискать релевантной информации, но нашел только вой и плачь пострадавших.

Судя по всему, тулза писалась ребятами из mail.ru под какие-то свои узкоспециализированные цели и под определенное
окружение. Задорная идея — залезть в кишки репликатора и поковырять там палкой в сишном коде у меня вспыхнула, но
уперлась в отсутствующее время, и потухла.

Тогда было решено последовать примеру великих:

![my-replicator](my-own-replicator.png)

Итак, встречайте, [очередная узкоспециализированная тулза, для репликации из мускула в tarantool](../../replicator)!

Читает binlog и формирует результаты в запросы к тарантулу. Последнюю позицию старается фиксировать в локальный файл.

<a name="conf-tarantool"></a>

### Tarantool

Скрипт инициализации:

```lua
box.cfg {
    listen = 3301,
    log_level = 2,
    net_msg_max = 7680,
    memtx_memory = 536870912
}

box.schema.space.create('users', { if_not_exists = true }):format({
    { name = 'id', type = 'unsigned' },
    { name = 'username', type = 'string' },
    { name = 'first_name', type = 'string' },
    { name = 'last_name', type = 'string' },
    { name = 'age', type = 'unsigned' },
    { name = 'gender', type = 'string' },
    { name = 'city', type = 'string' },
    { name = 'password', type = 'string' },
    { name = 'created_at', type = 'string' },
    { name = 'interests', type = 'string' },
})
box.space.users:create_index('primary', { type = "hash", unique = true, parts = { 'id' }, if_not_exists = true })
box.space.users:create_index('first_name_idx', { type = 'TREE', unique = false, parts = { 'first_name' }, if_not_exists = true })
box.schema.user.grant('spiderman', 'read,write,execute', 'space', 'users')

dofile('searcher.lua')
```

Скрипт с функцией поиска:

```lua
function search(prefix_first_name, prefix_second_name, offset, limit)
    local tuples = box.space.users.index.first_name_idx:select(prefix_first_name, { iterator = 'GE', offset = offset })
    local count = 0
    local results = {}
    for _, tuple in ipairs(tuples) do
        if count >= limit then
            return results
        end
        if string.startswith(tuple[3], prefix_first_name, 1, -1) and string.startswith(tuple[4], prefix_second_name, 1, -1) then
            table.insert(results, tuple)
            count = count + 1
        end
    end
    return results
end
```

<a name="conf-mysql"></a>

### Mysql

В [конфиг](../../deployment/tarantool/mysql/master.cnf) обязательно добавляем:

- binlog_format = ROW
- server-id=1

Так же создается пользователь, который умеет дампить базу и читать binlog:

```mysql
  CREATE USER '$REPL_USER'@'%' IDENTIFIED BY '$REPL_PASSWORD';
GRANT REPLICATION CLIENT, REPLICATION SLAVE, RELOAD, PROCESS ON *.* TO '$REPL_USER'@'%';
GRANT SELECT ON `otus_ha`.* TO '$REPL_USER'@'%';
FLUSH PRIVILEGES;
```

<a name="load-tests"></a>

## Нагрузочное тестирование

<a name="tests-what"></a>

### Что тестируем

Тестировать будем поиск пользователей по префиксам имени и фамилии.

В API это роут вида
> `GET` /v1/users?lastName=[string]&firstName=[string]&limit=[int]

Запрос к MySQL будет выглядеть примерно так:

```mysql
SELECT u.id,
       u.username,
       u.first_name,
       u.last_name,
       u.age,
       u.gender,
       u.city,
       u.password_hash,
       u.created_at,
       u.interests
FROM users u
WHERE first_name = ?
  AND last_name = ?
LIMIT ?, ?;
```

<a name="tests-how"></a>

### Чем и как тестируем

Тесты для сборок с MySQL и Tarantool проводились отдельно.

Запросы осуществлялись утилитой `wrk`, в 10 потоков, по три раунда в одну минуту.

Раунды по 10, 100 и 1000 одновременных соединений.

Для формирования запросов к API `wrk` задействует lua-скрипт:

```lua
local charset = {}
do
    -- [a-z]
    for c = 97, 122 do
        table.insert(charset, string.char(c))
    end
end

local function randomString(length)
    if not length or length <= 0 then
        return ''
    end
    math.randomseed(os.clock() ^ 5)
    return randomString(length - 1) .. charset[math.random(1, #charset)]
end

request = function()
    local lastName = randomString(2)
    local firstName = randomString(2)

    path = "/v1/users?lastName=" .. lastName .. "&firstName=" .. firstName .. "&limit=3"
    return wrk.format("GET", path)
end
```

<a name="tests-prepearings"></a>

### Подготовка

**Соединимся с MySQL**

Проверим индексы:

```
> show index from users;
+-------+------------+----------------------+--------------+-------------+-----------+-------------+----------+--------+------+------------+
| Table | Non_unique | Key_name             | Seq_in_index | Column_name | Collation | Cardinality | Sub_part | Packed | Null | Index_type |
+-------+------------+----------------------+--------------+-------------+-----------+-------------+----------+--------+------+------------+
| users |          0 | PRIMARY              |            1 | id          | A         |     1018446 |     NULL | NULL   |      | BTREE      |
| users |          0 | un_idx               |            1 | username    | A         |      865312 |     NULL | NULL   |      | BTREE      |
| users |          1 | f_name_idx           |            1 | first_name  | A         |        5061 |     NULL | NULL   | YES  | BTREE      |
+-------+------------+----------------------+--------------+-------------+-----------+-------------+----------+--------+------+------------+
```

Проверим количество записей:

```
> select count(*) from users;
+----------+
| count(*) |
+----------+
|  1000002 |
+----------+
```

**Соединимся с Tarantool**,
> sudo docker exec -it tarantool console

Проверим индексы:

```
unix/:/var/run/tarantool/tarantool.sock> box.space.users.index
---
- 0: &0
    unique: true
    parts:
    - type: unsigned
      is_nullable: false
      fieldno: 1
    type: HASH
    id: 0
    space_id: 512
    name: primary
  1: &1
    unique: false
    parts:
    - type: string
      is_nullable: false
      fieldno: 3
    id: 1
    space_id: 512
    type: TREE
    name: first_name_idx
  primary: *0
  first_name_idx: *1
...
```

Проверим количество записей:

```
unix/:/var/run/tarantool/tarantool.sock> box.space.users:count()
---
- 1000002
...
```

Проверим процедуру поиска:

```
 search('Sl', 'Vo', 0, 3)
---
- - [2, 'madscie', 'Sludge', 'Vohaul', 125, 'm', 'Hidden Space Base', '$2a$10$sS1EKXztHWywwsQr5xCERe92goE2UIUuOXF.yrabdH1aRGxbIx2J.',
    '2021-03-03 17:28:00', 'Experiments, Evil Plans, Conquer the Galaxy']
...
```

Итак, репликация работает, данные в тарантуле есть и доступны процедуре поиска.

> Причину выбора именно таких индексов можно подглядеть в [отчете по ДЗ "Индексы"](../indexes/report.md)

**Создаем учётку тестера**

```
curl -XPOST http://127.0.0.1:8007/v1/auth/sign-up -d '{"username":"tester", "password":"1234567890", "passwordConfirm":"1234567890", "gender":"f"}'
```

Из ответа берем авторизационный токен.

<a name="testing"></a>

### Прогон тестов

<a name="tests-mysql"></a>

#### Mysql

Раунд 1

```
wrk -t10 -c10 -d1m --timeout 30s -H "Authorization: Bearer: eyJhb ... zfYAg" -s ./request.lua http://127.0.0.1:8007

Running 1m test @ http://127.0.0.1:8007
  10 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   295.24ms  469.95ms   3.66s    86.96%
    Req/Sec    37.16     49.44   320.00     86.17%
  8811 requests in 1.00m, 3.17MB read
Requests/sec:    146.68
Transfer/sec:     54.10KB
```

Раунд 2

```
wrk -t10 -c100 -d1m --timeout 30s -H "Authorization: Bearer: eyJhb ... zfYAg" -s ./request.lua http://127.0.0.1:8007

Running 1m test @ http://127.0.0.1:8007
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   932.63ms  966.66ms   7.10s    85.95%
    Req/Sec    18.63     17.78   222.00     85.88%
  7926 requests in 1.00m, 2.85MB read
  Socket errors: connect 0, read 1, write 0, timeout 0
Requests/sec:    131.96
Transfer/sec:     48.58KB
```

Раунд 3

```
wrk -t10 -c1000 -d1m --timeout 30s -H "Authorization: Bearer: eyJhb ... zfYAg" -s ./request.lua http://127.0.0.1:8007

Running 1m test @ http://127.0.0.1:8007
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     3.97s     2.44s   10.00s    65.42%
    Req/Sec    15.99     14.09   130.00     88.25%
  6652 requests in 1.00m, 2.38MB read
  Socket errors: connect 0, read 1347, write 0, timeout 0
Requests/sec:    110.73
Transfer/sec:     40.53KB
```

<a name="tests-tarantool"></a>

#### Tarantool

Раунд 1

```
wrk -t10 -c10 -d1m --timeout 30s -H "Authorization: Bearer: eyJhb ... zfYAg" -s ./request.lua http://127.0.0.1:8007

Running 1m test @ http://127.0.0.1:8007
  10 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.10s     1.95s    9.29s    63.96%
    Req/Sec     0.01      0.09     1.00     99.10%
  111 requests in 1.00m, 39.48KB read
Requests/sec:      1.85
Transfer/sec:     672.79B
```

Раунд 2

```
wrk -t10 -c100 -d1m --timeout 30s -H "Authorization: Bearer: eyJhb ... zfYAg" -s ./request.lua http://127.0.0.1:8007

Running 1m test @ http://127.0.0.1:8007
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.07ms   14.63ms 399.85ms   99.37%
    Req/Sec     3.21k   367.96     5.64k    78.56%
  951206 requests in 1.00m, 351.06MB read
  Socket errors: connect 0, read 100, write 0, timeout 0
  Non-2xx or 3xx responses: 951204
Requests/sec:  15833.62
Transfer/sec:      5.84MB
```

Раунд 3

```
wrk -t10 -c1000 -d1m --timeout 30s -H "Authorization: Bearer: eyJhb ... zfYAg" -s ./request.lua http://127.0.0.1:8007

Running 1m test @ http://127.0.0.1:8007
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    24.56ms   32.14ms 259.02ms   81.43%
    Req/Sec     4.07k   538.03     5.38k    76.01%
  805136 requests in 1.00m, 297.15MB read
  Socket errors: connect 0, read 2000, write 0, timeout 0
  Non-2xx or 3xx responses: 805136
Requests/sec:  13406.51
Transfer/sec:      4.95MB
```

<a name="pivot-tables"></a>

### Сводные таблицы

**10 потоков по 10 соединений**

| |MySQL|Tarantool|
|-----|------|------|
|Latency|295.24ms|5.10s|
|Throughput|54.10KB|672.79B|
|RPS|146.68|1.85|
|Errors|0|0|

В сравнении мускулом очень вяло. О возможных причинах в конце секции.

**10 потоков по 100 соединений**

| |MySQL|Tarantool|
|-----|------|------|
|Latency|932.63ms|4.07ms|
|Throughput|48.58KB|5.84MB|
|RPS|131.96|15833.62|
|Errors|0|951204|

Тарантул показал отличные результаты только по той причине, что прилег на бок, и API отвечало ошибкой с минимальными
задержками. Это доказывает большое количество ошибок в результатах.

В лог посыпали шквалом ошибки следующего содержания:
> Internal Error: client connection is not ready (0x4000)

**10 потоков по 1000 соединений**

| |MySQL|Tarantool|
|-----|------|------|
|Latency|3.97s|24.56ms|
|Throughput|40.53KB|4.95MB|
|RPS|110.73|13406.51|
|Errors|0|805136|

Тарантул окончательно захлебнулся, перестав реагировать на внешние раздражители, тычки палкой и попытки соединиться с
консолью.

<a name="tests-results"></a>

### Мысли по результатам

Похоже, что под большим объемом запросов мы выгребли некую очередь обработки запросов, после чего тарантул начал рубить
запросы.

Попытки поиграться с настройками и выкручивание параметров на максимум к улучшению показателя отказов не привели.

Судя по всему очень сильно нагнетал поисковый lua скрипт, который проводил вторичный поиск путем обхода массива с
неупорядоченными данными, полученными в результате первичного поиска по индексу. Сложность такого подхода - величина
линейная, зависящая от количества полученных при первой выборке кортежей. Попытки оптимизировать поисковый lua скрипт
ничем не закончились, возможно тут виной моя компетенция в lua,- язык для меня новый.

Думаю, что ситуация выглядела бы на много лучше, при запуске нескольких реплицированных инстансов тарантула, но на
данный эксперимент времени уже не осталось.

Итого: чуда не случилось, мускул забодал паука на несвойственной тому задаче.

<a name="total"></a>

## Выводы

Неудачно был функционал для работы с tarantool, но лень и экономия времени предопределили выбор, т.к. для функционала
поиска пользователей ранее уже выполнялось нагрузочное тестирование, и под это были готовы как окружение, так и сеятель
данных.

Как я понимаю, тарантул хорош в качестве key-value хранилища данных. Возможно, он даст жары там, где необходима пред или
пост обработка данных на записи/чтении, или какая-то вариация map-reduce. Думаю, он отлично бы себя проявил, как бд для
хранения и первичной обработки данных, поступающих от потоковых систем. Но вот на запросах вида LIKE ему уж очень
неуютно, что не удивительно,- на большом объеме неотсортированных данных выполнять поиск перебором — не самая крутая
идея.

Были некие иллюзии, что за счет оперирования исключительно памятью, Тарантул зарешает мускул даже на такой
несвойственной ему задаче, но увы, этого не произошло.

На будущее, надо столкнуть Tarantool с Redis, на какой-то более свойственной данному типу хранилищ задаче. Учитывая, что
они оба умеют в lua, это будет занимательный эксперимент, с неплохим поводом для набросов на вентилятор по результатам.

По подгоранию my metal ass. Для меня осталась непонятой ситуация с репликатором от mail.ru. Возможно стоит пересмотреть
условия задачи, т.к. данный проект выглядит заброшенным, имеет вагон проблем и оч недружественные требования, при том
движений в официальном репозитории особо не видно.

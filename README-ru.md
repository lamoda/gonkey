# Gonkey: инструмент автоматизации тестирования

Gonkey протестирует ваши сервисы, используя их API. Он умеет обстреливать сервис заранее заготовленными запросами и проверять ответы. Сценарий теста описывается в YAML-файле.

Возможности:

- работает с REST/JSON API
- проверка API сервиса на соответствие OpenAPI-спеке
- заполнение БД сервиса данными из фикстур (поддерживается PostgreSQL, MySQL, Aerospike, Redis)
- моки для имитации внешних сервисов
- можно подключить к проекту как библиотеку и запускать вместе с юнит-тестами
- запись результата тестов в виде отчета [Allure](http://allure.qatools.ru/)
- имеется [JSON-schema](#json-schema) для автодополнения и валидации YAML-файлов Gonkey

## Содержание

- [Использование консольной утилиты](#использование-консольной-утилиты)
- [Использование gonkey как библиотеки](#использование-gonkey-как-библиотеки)
- [Пример тестового сценария](#пример-тестового-сценария)
- [Статус теста](#статус-теста)
- [HTTP-запрос](#http-запрос)
- [HTTP-ответ](#http-ответ)
- [Переменные](#переменные)
  - [Способы присвоения](#способы-присвоения)
    - [В описании самого теста](#в-описании-самого-теста)
    - [Из результатов предыдущего запроса](#из-результатов-предыдущего-запроса)
    - [Из результата текущего запроса](#из-результата-текущего-запроса)
    - [В переменных окружения или в env-файле](#в-переменных-окружения-или-в-env-файле)
    - [В cases](#в-cases)
- [Загрузка файлов](#загрузка-файлов)
- [Фикстуры](#фикстуры)
  - [Удаление данных из таблиц](#удаление-данных-из-таблиц)
  - [Шаблоны записей](#шаблоны-записей)
  - [Наследование записей](#наследование-записей)
  - [Связывание записей](#связывание-записей)
  - [Выражения](#выражения)
  - [Aerospike](#aerospike)
  - [Redis](#redis)
- [Моки](#моки)
  - [Запуск моков при использовании gonkey как библиотеки](#запуск-моков-при-использовании-gonkey-как-библиотеки)
  - [Описание моков в файле с тестом](#описание-моков-в-файле-с-тестом)
    - [Проверки запросов (requestConstraints)](#проверки-запросов-requestconstraints)
    - [Стратегии ответов (strategy)](#стратегии-ответов-strategy)
    - [Подсчет количества вызовов](#подсчет-количества-вызовов)
- [Использование shell скриптов](#использование-shell-скриптов)
  - [Описание скрипта](#описание-скрипта)
  - [Запуск скрипта с параметризацией](#запуск-скрипта-с-параметризацией)
- [Запрос в Базу данных](#запрос-в-базу-данных)
  - [Формат описания запросов](#формат-описания-запросов)
  - [Описание запроса](#описание-запроса)
  - [Описание ответа на запрос в Базу данных](#описание-ответа-на-запрос-в-базу-данных)
  - [Параметризация при запросах в Базу данных](#параметризация-при-запросах-в-базу-данных)
  - [Игнорирование порядка записей в ответе на запрос в базу данных](#игнорирование-порядка-записей-в-ответе-на-запрос-в-базу-данных)
- [JSON-schema](#json-schema)
  - [Настройка на IDE Jetbrains](#настройка-на-ide-jetbrains)
  - [Настройка на IDE VSCode](#настройка-на-ide-vscode)

## Использование консольной утилиты

Для тестирование сервиса, размещенного на удаленном хосте, используйте gonkey как консольную утилиту.

`./gonkey -host <...> -tests <...> [-spec <...>] [-db_dsn <...> -fixtures <...>] [-allure] [-v]`

- `-spec <...>` путь к файлу или URL со swagger-спецификацией сервиса
- `-host <...>` хост:порт сервиса
- `-tests <...>` файл или директория с тестами
- `-db-type <...>` - тип базы данных. В данный момент поддерживается PostgreSQL, Aerospike, Redis.
- `-db_dsn <...>` dsn для вашей тестовой SQL базы данных (бд будет очищена перед наполнением!), поддерживается только PostgreSQL
- `-aerospike_host <...>` при использовании Aerospike - URL для подключения к нему в формате `host:port/namespace`
- `-redis_url <...>` при использовании Redis - адрес для подключения к Redis, например `redis://user:password@localhost:6789/1?dial_timeout=1&db=1&read_timeout=6s&max_retries=2`
- `-fixtures <...>` директория с вашими фикстурами
- `-allure` генерировать allure-отчет
- `-v` подробный вывод
- `-debug` отладочный вывод

В таком режиме моки использовать не получится.

## Использование gonkey как библиотеки

Чтобы интегрировать функциональные тесты в нативные тесты Go и запускать их вместе, используйте gonkey как библиотеку.

Создайте файл для будущего теста, например, `func_test.go`.

Подключите gonkey как зависимость к вашему проекту в этом файле.

```go
import (
    "github.com/lamoda/gonkey/runner"
    "github.com/lamoda/gonkey/mocks"
)
```

Создайте функцию с тестом.

```go
package test

import (
  "testing"

  "github.com/lamoda/gonkey/fixtures"
  "github.com/lamoda/gonkey/mocks"
  "github.com/lamoda/gonkey/runner"
)

func TestFuncCases(t *testing.T) {
  // проинициализируйте моки, если нужно (подробнее - ниже)
  // m := mocks.NewNop(...)

  // проинициализируйте базу для загрузки фикстур, если нужно (подробнее - ниже)
  // db := ...

  // проинициализируйте Aerospike для загрузки фикстур, если нужно (подробнее - ниже)
  // aerospikeClient := ...

  // создайте экземпляр сервера вашего приложения
  srv := server.NewServer()
  defer srv.Close()

  // запустите выполнение тестов из директории cases с записью в отчет Allure
  runner.RunWithTesting(t, &runner.RunWithTestingParams{
    Server:   srv,
    TestsDir: "cases",
    Mocks:    m,
    DB:       db,
    Aerospike: runner.Aerospike{
      Client:    aerospikeClient,
      Namespace: "test",
    },
    // Тип используемой базы данных, возможные значения fixtures.Postgres, fixtures.Mysql, fixtures.Aerospike, fixtures.CustomLoader
    // Если в параметр DB не пустой, а данный параметр не назначен, будет использоваться тип бд fixtures.Postgresql
    DbType:      fixtures.Postgres,
    FixturesDir: "fixtures",
  })
}
```

Начиная с версии 1.18.3, добавлена поддержка внешних модулей для загрузки тестовых данных из фикстур, если gonkey используется как библиотека.
Чтобы начать использовать внешний загрузчик, вы должны импортировать модуль, содержащий реализацию интерфейса fixtures.Loader.

Пример для загрузки данных в Redis

```go
package test

import (
  "net/http"
  "net/http/httptest"
  "testing"

  "github.com/lamoda/gonkey/fixtures"
  redisLoader "github.com/lamoda/gonkey/fixtures/redis"
  // redisLoader "custom_module/gonkey-redis" // внешняя библиотека, содержащая реализацию интерфейса fixtures.Loader
  redisClient "github.com/go-redis/redis/v9"
  "github.com/lamoda/gonkey/runner"
)

func TestFuncCases(t *testing.T) {
  serveMux := http.NewServeMux()
  
  serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    _, _ = w.Write([]byte("ok"))
  })
  
  srv := httptest.NewServer(serveMux)
  
  clientOptions, err := redisClient.ParseURL("redis://user:password@localhost:6789/1?dial_timeout=1&db=1&read_timeout=6s&max_retries=2")
  if err != nil {
      panic(err)
  }

  redisFixtureLoader := redisLoader.New(redisLoader.LoaderOptions{
    FixtureDir: "./fixtures",
    Redis:      clientOptions,
  })

  runner.RunWithTesting(t, &runner.RunWithTestingParams{
    Server:        srv,
    TestsDir:      "./cases",
    DbType:        fixtures.CustomLoader,
    FixtureLoader: redisFixtureLoader,
  })
}
```

Теперь тесты можно запускать через `go test`, например, так: `go test ./...`.

## Пример тестового сценария

```yaml
- name: КОГДА запрашивается список заказов ДОЛЖЕН успешно возвращаться
  method: GET
  status: ""
  path: /jsonrpc/v2/order.getBriefList
  query: ?id=550e8400-e29b-41d4-a716-446655440000&jsonrpc=2.0&user_id=00001

  fixtures:
    - order_0001
    - order_0002

  response:
    200: |
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "jsonrpc": "2.0",
        "result": {
          "data": [
            "ORDER0001",
            "ORDER0002"
          ],
          "meta": {
            "items": 0,
            "limit": 50,
            "page": 0,
            "pages": 0
          }
        }
      }

- name: КОГДА запрашивается один заказ ДОЛЖЕН возвращаться пользователь и сумма
  method: POST
  path: /jsonrpc/v2/order.getOrder

  headers:
    Authorization: Bearer HsHG67d38hJKJFdfjj==
    Content-Type: application/json

  cookies:
    sid: ZmEwZDkwYzgwMmQzMGIzOGIxODM3ZmFiOTGJhMzU=
    lid: AAAEAFu/TdhHBg7UAgA=

  comparisonParams:
    ignoreValues: false
    ignoreArraysOrdering: false
    disallowExtraFields: false

  request: |
    {
      "jsonrpc": "2.0",
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "method": "order.getOrder",
      "params": [
        {
          "order_nr": {{ .orderNr }}
        }
      ]
    }

  response:
    200: |
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "jsonrpc": "2.0",
        "result": {
          "user_id": {{ .userId }},
          "amount": {{ .amount }}
        }
      }

responseHeaders:
    200:
      Content-Type: "application/json"
      Cache-Control: "no-store, must-revalidate"
      Set-Cookie: "mycookie=123; Path=/; Domain=mydomain.com", "mycookie=456; Path=/; Domain=.mydomain.com"

  cases:
    - requestArgs:
        orderNr: ORDER0001
      responseArgs:
        200:
          userId: '0001'
          amount: 1000

    - requestArgs:
        orderNr: ORDER0002
      responseArgs:
        200:
          userId: '0001'
          amount: 72000
```

Как видно из примера, вы можете использовать Regexp для проверки тела ответа.
Они могут быть использованы для проверки всего тела (если это просто текст):

```yaml
    response:
        200: "$matchRegexp(^xy+z$)"
```

или для элементов map/array (если это JSON):

```yaml
    response:
        200: |
          {
            "id": "$matchRegexp([\\w-]+)",
            "jsonrpc": "$matchRegexp([12].0)",
            "result": [
              "data": [
                  "$matchRegexp(ORDER[0]{3}[0-9])",
                  "$matchRegexp(ORDER[0]{3}[0-9])"
              ],
            ]
          }
```

Так же в поле query вначале указывать "?" необязательно

## Статус теста

`status` - параметр, для того чтобы помечать тесты, может иметь следующие значения:

- `broken` - такой тест не будет запущен, в отчете будет отмечен как `broken`
- `skipped` - такой тест не будет запущен, в отчете будет отмечен как `skipped`
- `focus` - если у теста выставлен такой статус, все остальные тесты в suite у которых не проставлен статус, будут отмечены как `skipped` и будут запущены только тесты с статусом `focus`

## HTTP-запрос

`method` - параметр для передачи типа HTTP запроса, формат передачи указан в примере выше

`path` - параметр для передачи URL-пути, формат передачи указан в примере выше

`headers` - параметр для передачи http-заголовков, формат передачи указан в примере выше.

`cookies` -  параметр для передачи cookie, формат передачи указан в примере выше.

## HTTP-ответ

`response` - тело ответа HTTP для указанных кодов состояния HTTP.

`responseHeaders` - все заголовки ответа HTTP для указанных кодов состояния HTTP.

## Переменные

В описании теста можно использовать переменные, они поддерживаются в следующих полях:

- method
- description
- path
- query
- headers
- request
- response
- dbQuery
- dbResponse
- body для моков
- headers для моков
- requestConstraints для моков

Пример использования:

```yaml
- method: "{{ $method }}"
  description: "{{ $description }}"
  path: "/some/path/{{ $pathPart }}"
  query: "{{ $query }}"
  headers:
    header1: "{{ $header }}"
  request: '{"reqParam": "{{ $reqParam }}"}'
  response:
    200: "{{ $resp }}"
  mocks:
    server_mock:
      strategy: constant
      body: >
        {
          "message": "{{ $mockParam }}"
        }
      statusCode: 200
  dbQuery: >
    SELECT id, name FROM testing_tools WHERE id={{ $sqlQueryParam }}
  dbResponse:
    - '{"id": {{ $sqlResultParam }}, "name": "gonkey"}'
```

Присваивать значения переменным можно следующими способами:

- в описании самого теста
- из результатов предыдущего запроса
- из результата текущего запроса
- в переменных окружения или в env-файле

Приоритеты источников соответствуют порядку перечисления.

### Способы присвоения

#### В описании самого теста

Пример:

```yaml
- method: "{{ $method }}"
  path: "/some/path/{{ $pathPart }}"
  variables:
    reqParam: "reqParam_value"
    method: "POST"
    pathPart: "part_of_path"
    query: "query_val"
    header: "header_val"
    resp: "resp_val"
  query: "{{ $query }}"
  headers:
    header1: "{{ $header }}"
  request: '{"reqParam": "{{ $reqParam }}"}'
  response:
    200: "{{ $resp }}"
```

#### Из результатов предыдущего запроса

Пример:

```yaml
# если в ответе plain text
- name: "get_last_post_id"
  ...
  variables_to_set:
          200: "id"

# если в ответе JSON
- name: "get_last_post_info"
  variables_to_set:
          200:
            id: "id"
            title: "title"
            authorId: "author_info.id"
```

Обратите внимание - если нужно использовать значение вложенного поля, можно указать путь до него:
> "author_info.id"

Глубина вложенности может быть любая.

#### Из результата текущего запроса

Пример:

```yaml
- name: Get info with database
  method: GET
  path: "/info/1"
  variables_to_set:
    200:
      golang_id: query_result.0.0
  response:
    200: '{"result_id": "1", "query_result": [[ {{ $golang_id }} , "golang"], [2, "gonkey"]]}'
  dbQuery: >
    SELECT id, name FROM testing_tools WHERE id={{ $golang_id }}
  dbResponse:
    - '{"id": {{ $golang_id}}, "name": "golang"}'
```

#### В переменных окружения или в env-файле

Gonkey автоматически проверяет наличие указанной переменной среди переменных окружения (в таком же регистре) и берет значение оттуда, в случае наличия.

Если указан env-файл, то описанные в нем переменные добавятся/заменят соответствующие перемнные окружения.
env-файл указывается с помощью параметра env-file

Пример env-файла (стандартный синтаксис):

```.env
jwt=some_jwt_value
secret=my_secret
password=private_password
```

env-файл, например, удобно использовать, когда нужно вынести из теста приватную информацию (пароли, ключи и т.п.)

#### В cases

Переменные могут быть заданы в блоке *cases*.

Пример:

```yaml
- name: Get user info
  method: GET
  path: "/user/1"
  response:
    200: '{ "user_id": "1", "name": "{{ $name }}", "surname": "{{ $surname }}" }'
  cases:
    - variables:
        name: John
        surname: Doe
```

Такие переменные будут доступны и в других кейсах, если не будут переопределены.

## Загрузка файлов

В тестовом запросе можно загружать файлы. Для этого нужно указать тип запроса - POST и заголовок:

> Content-Type: multipart/form-data

Пример:

```yaml
 - name: "upload-files"
   method: POST
   form:
       files:
         file1: "testdata/upload-files/file1.txt"
         file2: "testdata/upload-files/file2.log"
   headers:
     Content-Type: multipart/form-data # case-sensitive, can be omitted
   response:
     200: |
       {
         "status": "OK"
       }
```

## Фикстуры

Чтобы наполнить базу перед тестом, используются файлы с фикстурами.

Пример файла:

```yaml
# fixtures/comments.yml
inherits:
  - another_fixture
  - yet_another_fixture

tables:
  posts:
    - $name: janes_post
      title: New post
      text: Post text
      author: Jane Dow
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

    - $name: apples_post
      title: Morning digest
      text: Text
      author: Apple Seed
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

  comments:
    - post_id: $janes_post.id
      content: A comment...
      author_name: John Doe
      author_email: john@doe.com
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

    - post_id: $apples_post.id
      content: Another comment...
      author_name: John Doe
      author_email: john@doe.com
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

  another_table:
      ...
  ...
```

Записи в фикстурах можно наследовать одну от другой, использовать шаблоны, а так же ссылаться из одной записи на другую.

### Удаление данных из таблиц

Чтобы очистить таблицу перед тестом, укажите пару квадратных скобок рядом с названием таблицы в файле с фикстурами.

Пример:

```yaml
# fixtures/empty_posts_table.yml
tables:
  posts: []
```

### Шаблоны записей

Обычно, чтобы вставить строку данных в базу, вам нужно перечислить все поля, для которых в базе не предусмотрено значение по умолчанию. Довольно часто, многие из этих полей не важны для теста и их значения повторяются от одной фикстуры к другой, создавая ненужный визуальный мусор и усложняя их поддержку.

С помощью шаблонов вы можете наследовать поля из шаблонной записи, каждый раз переопределяя только те поля, которые действительно важны для теста.

Пример определения шаблона:

```yaml
templates:
  dummy_client:
    name: Dummy Client Name
    age: 35
    ip: 127.0.0.1
    is_deleted: false

  dummy_deleted_client:
    $extend: dummy_client
    is_deleted: true

tables:
  ...
```

Пример использования шаблона в фикстуре:

```yaml
templates:
   ...
tables:
   clients:
      - $extend: dummy_client
      - $extend: dummy_client
        name: Josh
      - $extend: dummy_deleted_client
        name: Jane
```

Как вы могли заметить, шаблоны тоже можно наследовать друг от друга с помощью ключевого слова `$extend`, но только если к моменту определения зависимого шаблона базовый шаблон уже определен (в этом же файле или любом, подключенном через `inherits`).

### Наследование записей

Записи, как и шаблоны, можно наследовать с помощью `$extend`.

Для того, чтобы унаследовать запись, сначала нужно присвоить строке имя с помощью `$name`:

```yaml
# fixtures/post.yaml
tables:
  posts:
    - $name: regular_post
      title: Post title
      text: Some text
```

Имена, присваеваемые строкам, должны быть уникальны среди всех загружаемых файлов с фикстурами, а также не пересекаться с именами шаблонов.

В другом файле фикстур нужно объявить, что определенная строка наследует ранее объявленную запись с помощью `$extend`, так же, как и в случае с шаблоном:

```yaml
# fixtures/deleted_post.yaml
inherits:
  - post
tables:
  posts:
    - $extend: regular_post
      is_deleted: true
```

Не забудьте объявить зависимость между файлами в `inherits`, чтобы один файл всегда загружался вместе с другим.

Обратите внимание, что наследование строк работает только между разными файлами фикстур. В пределах одного файла объявить наследование невозможно.

### Связывание записей

Несмотря на то, что файлы фикстур позволяют вам задавать значение для автоинкрементных колонок (обычно `id`), не рекомендуется этого делать. Контролировать, чтобы в разных файлах использовались правильные значения для `id`, и чтобы они нигде не пересеклись, очень сложно. Чтобы база сама присвоила значение автоинкрементному полю, достаточно просто не указывать его значение явно.

Однако, если не указывать значение для `id`, то как связать несколько сущностей, которые должны ссылаться друг на друга по идентификаторам? Фикстуры позволяют ссылаться на значение ранее вставленных в базу строк по их имени, используя нотацию `$refName.fieldName`.

Объявим именованную запись:

```yaml
# fixtures/post.yaml
tables:
  posts:
    - $name: regular_post
      title: Post title
      text: Some text
```

Теперь, чтобы связать таблицы `posts` и `comments`, обратимся к записи по имени (`$regular_post`) и укажем поле, из которого следует взять значение (`id`):

```yaml
# fixtures/comment.yaml
tables:
  comments:
    - post_id: $regular_post.id
      content: A comment...
      author_name: John Doe
```

Ссылаться можно только на поля ранее вставленной в базу записи, на поля шаблона ссылаться нельзя, при попытке это сделать, вы получите ошибку `undefined reference`.

Обратите внимание на ограничение: нельзя ссылаться на записи в пределах одной таблицы одного файла.

### Выражения

Если в базу нужно записать не статичное значение, а результат исполнения выражения, то можно воспользоваться конструкцией `$eval()`. Все, что будет задано внутри скобок, будет вставлено в базу в сыром, неэкранированном виде. Таким образом, внутри `$eval()` можно написать все то, что вы могли бы написать в самом запросе.

Например, такая конструкция вставит текущую дату и время в качестве значения поля:

```yaml
tables:
  comments:
    - created_at: $eval(NOW())
```

### Aerospike
Для хранилища Aerospike также поддерживается заливка тестовых данных. Для этого важно не забыть при запуске gonkey как CLI-приложение использовать флаг `-db-type aerospike`, а при использовании в качестве библиотеки в конфигурации раннера: `DbType: fixtures.Postgres`.

Формат файлов с фикстурами для Aerospike отличается, но смысл остаётся прежним:

```yaml
sets:
  set1:
    key1:
      bin1: "value1"
      bin2: 1
    key2:
      bin1: "value2"
      bin2: 2
      bin3: 2.569947773654566473
  set2:
    key1:
      bin4: false
      bin5: null
      bin1: '"'
    key2:
      bin1: "'"
      bin5:
        - 1
        - '2'
```

Также поддерживаются шаблоны:

```yaml
templates:
  base_tmpl:
    bin1: value1
  extended_tmpl:
    $extend: base_tmpl
    bin2: value2

sets:
  set1:
    key1:
      $extend: base_tmpl
      bin1: overwritten
  set2:
    key1:
      $extend: extended_tmpl
      bin2: overwritten
```

Связывание записей и выражения на данный момент не поддерживаются.

### Redis

Поддерживается загрузка тестовых данных через фикстуры для хранилища ключ/значение Redis

Список, поддерживаемых структур данных:

- Пара ключ/значение
- Set
- Hash
- List
- ZSet (sorted set)

Связывание записей и выражения на данный момент не поддерживаются.

Пример файла фикстуры:

```yaml
inherits:
  - template1
  - template2
  - other_fixture
templates:
  keys:
    - $name: parentKeyTemplate
      values:
        baseKey:
          expiration: 1s
          value: 1
    - $name: childKeyTemplate
      $extend: parentKeyTemplate
      values:
        otherKey:
          value: 2
  sets:
    - $name: parentSetTemplate
      expiration: 10s
      values:
        - value: a
    - $name: childSetTemplate
      $extend: parentSetTemplate
      values:
        - value: b
  hashes:
    - $name: parentHashTemplate
      values:
        - key: a
          value: 1
        - key: b
          value: 2
    - $name: childHashTemplate
      $extend: parentHashTemplate
      values:
        - key: c
          value: 3
        - key: d
          value: 4
  lists:
    - $name: parentListTemplate
      values:
        - value: 1
        - value: 2
    - $name: childListTemplate
      values:
        - value: 3
        - value: 4
  zsets:
    - $name: parentZSetTemplate
      values:
        - value: 1
          score: 2.1
        - value: 2
          score: 4.3
    - $name: childZSetTemplate
      value:
        - value: 3
          score: 6.5
        - value: 4
          score: 8.7
databases:
  1:
    keys:
      $extend: childKeyTemplate
      values:
        key1:
          value: value1
        key2:
          expiration: 10s
          value: value2
    sets:
      values:
        set1:
          $extend: childSetTemplate
          expiration: 10s
          values:
            - value: a
            - value: b
        set3:
          expiration: 5s
          values:
            - value: x
            - value: y
    hashes:
      values:
        map1:
          $extend: childHashTemplate
          values:
            - key: a
              value: 1
            - key: b
              value: 2
        map2:
          values:
            - key: c
              value: 3
            - key: d
              value: 4
    lists:
      values:
        list1:
          $extend: childListTemplate
          values:
            - value: 1
            - value: 100
            - value: 200
    zsets:
      values:
        zset1:
          $extend: childZSetTemplate
          values:
            - value: 5
              score: 10.1
  2:
    keys:
      values:
        key3:
          value: value3
        key4:
          expiration: 5s
          value: value4
```

## Моки

Чтобы для тестов имитировать ответы от внешних сервисов, применяются моки.

Один мок - это поднятый "на лету" веб-сервер, который перед запуском каждого теста наполняется определенной логикой. Логика определяет, что ответит сервер на тот или иной запрос. Логика ответов описывается в файле теста.

### Запуск моков при использовании gonkey как библиотеки

Перед запуском тестов происходит старт всех планируемых к использованию моков - то есть поднимается заданное количество серверов, для каждого из них выделяется случайный порт.

```go
// создаем пустые моки сервисов
m := mocks.NewNop(
    "cart",
    "loyalty",
    "catalog",
    "madmin",
    "okz",
    "discounts",
)

// запускаем моки
err := m.Start()
if err != nil {
    t.Fatal(err)
}
defer m.Shutdown()
```

После того, как веб-серверы моков подняты, можно получить от них адреса (хост и порт), на которых они разместились. Используя эти адреса, вы конфигурируете свой сервис, чтобы вместо обращений к реальным системам он обращался к поднятым мок-серверам.

```go
// конфигурируем и запускаем наш сервис
srv := server.NewServer(&server.Config{
 CartAddr:      m.Service("cart").ServerAddr(),
 LoyaltyAddr:   m.Service("loyalty").ServerAddr(),
 CatalogAddr:   m.Service("catalog").ServerAddr(),
 MadminAddr:    m.Service("madmin").ServerAddr(),
 OkzAddr:       m.Service("okz").ServerAddr(),
 DiscountsAddr: m.Service("discounts").ServerAddr(),
})
defer srv.Close()
```

Как только вы подняли моки и сконфигурировали свой сервис, можно запускать тесты.

```go
runner.RunWithTesting(t, &runner.RunWithTestingParams{
    Server:    srv,
    Directory: "tests/cases",
    Mocks:     m, // передаем моки в раннер тестов
})
```

Дополнительно библиотека регистрирует специальные переменные окружения `GONKEY_MOCK_<ИМЯ_MOCK>`, которые содержат адрес и порт соответствующего мок-сервера. Эти переменные окружения вы можете использовать при написании тестов.

### Описание моков в файле с тестом

Каждый тест перед запуском сообщает мок-серверу конфигурацию, которая определяет, что мок-сервер ответит на тот или иной запрос. Эта конфигурация задается в YAML-файле с тестом в секции `mocks`.

Одновременно в файле с тестом можно описать любое количество мок-сервисов:

```yaml
- name: Test with mocks
  ...
  mocks:
    service1:
      ...
    service2:
      ...
    service3:
      ...
  request:
    ...
```

Описание каждого мок-сервиса состоит из:

`requestConstraints` - массив проверок, которые применяются к полученному запросу. Если хотя бы одна проверка не пройдена, тест считается проваленным. Список возможных проверок - ниже.

`strategy` - стратегия ответа мока на запросы. Список возможных стратегий - ниже.

Остальные ключи на первом уровне вложенности в описании мока - это параметры к стратегии. Их набор различен для каждой конкретной стратегии.

Пример конфигурации одного мок-сервиса:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - ...
        - ...
      strategy: strategyName
      strategyParam1: ...
      strategyParam2: ...
    ...
```

#### Проверки запросов (requestConstraints)

Запросы к мок-сервису можно валидировать с помощью одной или нескольких описанных ниже проверок.

Описание каждой проверки состоит из параметра `kind`, в котором указывается, что за проверка будет применена.

Все остальные ключи на этом уровне - это параметры проверки. У каждой проверки свой набор параметров.

##### nop

Пустая проверка. Всегда проходит успешно.

Нет параметров.

Пример:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: nop
    ...
```

##### bodyMatchesJSON

Проверяет, что тело запроса - это JSON, который соответствует заданному в параметре `body`.

Параметры:

- `body` (обязательный) - JSON, с которым будет сверяться запрос. Все ключи на всех уровнях, определенные в этом параметре, должны присутвовать в теле запроса.

Пример:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        # эта проверка будет требовать, чтобы запрос содержал ключи key1, key2 и subKey1,
        # а значения были равны value1 и value2. Однако в запросе допускаются другие ключи,
        # не перечисленные здесь - это нормально.
        - kind: bodyMatchesJSON
          body: >
            {
              "key1": "value1",
              "key2": {
                "subKey1": "value2",
              }
            }
    ...
```

##### bodyJSONFieldMatchesJSON

Когда тело запроса есть JSON, проверяет что конкретное поле этого JSON является строкой,
содержащей в себе другой JSON, соответствующий заданному.

Параметры:

- `path` (обязательный) - путь к JSON-полю, в котором находится проверяемое значение.
- `value` (обязательный) - JSON, с которым сверяется значение поля.

Пример:

Оригинальный запрос содержит вложенный JSON

```yaml
  {
      "field1": {
        "field2": "{\"wrapped\": \"json\"}"
      }
  }
```

Проверка вложенного JSON

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: bodyJSONFieldMatchesJSON
          path: field1.field2
          value: |
            {
              "wrapped": "json"
            }
  ...
```

##### pathMatches

Проверяет, что путь в запросе соответствует заданному значению.

Параметры:

- `path` - строка, которой должен быть равен путь запроса;
- `regexp` - регулярное выражение, которому должен соответствовать путь запроса.

Пример:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: pathMatches
          path: /api/v1/test/somevalue
    service2:
      requestConstraints:
        - kind: pathMatches
          regexp: ^/api/v1/test/.*$
    ...
```

##### queryMatches

Проверяет, что параметры GET запроса соответствуют заданным в параметре `query`.

Параметры:

- `expectedQuery` (обязательный) - строка параметров с которой будет сверяться запрос. Порядок параметров не имеет значения.

Пример:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        # эта проверка будет требовать, чтобы запрос содержал ключи key1 и key2,
        # а значения были равны key1=value1, key1=value11 и key2=value2. Ключи не указанные в запросе будут пропущены при проверке.
        - kind: queryMatches
          expectedQuery:  key1=value1&key2=value2&key1=value11
    ...
```

##### queryMatchesRegexp

Расширяет `queryMatches` так, чтобы можно было использовать проверку по регулярному выражению.

Параметры:

- `expectedQuery` (обязательный) - строка параметров с которой будет сверяться запрос. Порядок параметров не имеет значения.

Пример:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        # работает аналогично queryMatches с добавлением возможности вызова $matchRegexp
        - kind: queryMatchesRegexp
          expectedQuery:  key1=value1&key2=$matchRegexp(\\d+)&key1=value11
    ...
```

##### methodIs

Проверяет, что метод запроса соответствует заданному.

Параметры:

- `method` (обязательный) - строка, с которой сравнивается метод запроса.

Есть также два коротких варианта, не требущих указания параметра `method`:

- `methodIsGET`
- `methodIsPOST`

Примеры:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: methodIs
          method: PUT
    service2:
      requestConstraints:
        - kind: methodIsPOST
    ...
```

##### headerIs

Проверяет, что в запросе есть указанный заголовок и, опционально, что его значение равно заданному или подпадает под условия регулярного выражения.

Параметры:

- `header` (обязательный) - название заголовка, который ожидается в запросе;
- `value` - строка, которой должно быть равно значение заголовка;
- `regexp` - регулярное выражение, которому должно соответствовать значение заголовка.

Примеры:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: headerIs
          header: Content-Type
          value: application/json
    service2:
      requestConstraints:
        - kind: headerIs
          header: Content-Type
          regexp: ^(application/json|text/plain)$
    ...
```

##### bodyMatchesText

Проверяет что, тело запроса соответствует указанному или подпадает под условия регулярного выражения.

Параметры:

- `body` - строка, которой должно быть равно значение тела запроса;
- `regexp` - регулярное выражение, которому должно соответствовать значение тела.

Примеры:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: bodyMatchesText
            body: |-
              query HeroNameAndFriends {
                    hero {
                      name
                      friends {
                        name
                      }
                    }
                  }
    service2:
      requestConstraints:
        - kind: bodyMatchesText
            regexp: (HeroNameAndFriendsq)
    ...
```

##### bodyMatchesXML

Проверяет, что тело запроса - это XML, который соответствует заданному в параметре `body`.

Параметры:

- `body` (обязательный) - XML, с которым будет сверяться запрос.

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: bodyMatchesXML
          body: |
            <Person>
              <Company>Hogwarts School of Witchcraft and Wizardry</Company>
              <FullName>Harry Potter</FullName>
              <Email where="work">hpotter@hog.gb</Email>
              <Email where="home">hpotter@gmail.com</Email>
              <Addr>4 Privet Drive</Addr>
              <Group>
                <Value>Hexes</Value>
                <Value>Jinxes</Value>
                <Value>Jinxes</Value>
              </Group>
            </Person>
  ...
```

#### Стратегии ответов (strategy)

Стратегии ответов определяют, как мок будет отвечать на входящие запросы.

##### nop

Пустая стратегия. На любой запрос возвращается ответ `204 No Content` с пустым телом.

Не имеет параметров.

Пример:

```yaml
  ...
  mocks:
    service1:
      strategy: nop
    ...
```

##### file

Возвращает ответ, прочитанный из файла.

Параметры:

- `filename` (обязательный) - имя файла, из которого будет прочитано тело ответа;
- `statusCode` - HTTP-код ответа, по умолчанию `200`;
- `headers` - заголовки ответа.

Пример:

```yaml
  ...
  mocks:
    service1:
      strategy: file
      filename: responses/service1_success.json
      statusCode: 500
      headers:
        Content-Type: application/json
    ...
```

##### constant

Возвращает заданный ответ.

Параметры:

- `body` (обязательный) - задает тело ответа;
- `statusCode` - HTTP-код ответа, по умолчанию `200`;
- `headers` - заголовки ответа.

Пример:

```yaml
  ...
  mocks:
    service1:
      strategy: constant
      body: >
        {
          "status": "error",
          "errorCode": -32884,
          "errorMessage": "Internal error"
        }
      statusCode: 500
    ...
```

##### template

Стратегия дает возможность использовать в теле ответа, параметры входящего в мок запроса.
Реализовано с использованием пакета [text/template](https://pkg.go.dev/text/template).
Автоматически подгружает запрос в переменную `request`.

Параметры:

- `body` (обязательный) - задает тело ответа, должно быть совместимо с `text/template`;
- `statusCode` - HTTP-код ответа, по умолчанию `200`;
- `headers` - заголовки ответа.

Пример:

```yaml
  ...
  mocks:
    service1:
      strategy: template
      body: >
        {
          "value-from-query": {{ .request.Query "value" }},
          "data-from-body": {{ default 10 .request.Json.data }}
        }
      statusCode: 200
    ...
```

##### uriVary

Использует разные стратегии ответа, в зависимости от пути запрашиваемого ресурса.

При получении запроса на ресурс, который не задан в параметрах, тест считается проваленным.

Параметры:

- `uris` (обязательный) - список ресурсов, каждый ресурс можно сконфигурировать как отдельный мок-сервис, используя любые доступные проверки запросов и стратегии ответов (см. пример)
- `basePath` - общий базовый путь для всех ресурсов, по умолчанию пустой

Пример:

```yaml
  ...
  mocks:
    service1:
      strategy: uriVary
      basePath: /v2
      uris:
        /shelf/books:
          strategy: file
          filename: responses/books_list.json
          statusCode: 200
        /shelf/books/1:
          strategy: constant
          body: >
            {
              "error": "book not found"
            }
          statusCode: 404
    ...
```

##### methodVary

Использует разные стратегии ответа, в зависимости от метода запроса.

При получении запроса методом, который не упомянут в methodVary, тест считается проваленным.

Параметры:

- `methods` (обязательный) - список методов, каждый из которых можно сконфигурировать как отдельный мок-сервис, используя любые доступные проверки запросов и стратегии ответов (см. пример)

Пример:

```yaml
  ...
  mocks:
    service1:
      strategy: methodVary
      methods:
        GET:
          # ничего не мешает в этом месте использовать стратегию `uriVary`
          # тем самым можно формировать разные ответы на комбинацию метод+ресурс
          strategy: constant
          body: >
            {
              "error": "book not found"
            }
          statusCode: 404
        POST:
          strategy: nop
          statusCode: 204
    ...
```

##### sequence

На каждый последующий запрос эта стратегия будет отвечать так, как определено в очередной дочерней стратегии.

Если для запроса не задано дочерней стратегии, то есть пришло больше запросов, чем задано стратегий, то тест считается проваленным.

Параметры:

- `sequence` (обязательный) - список дочерних стратегий.

Пример:

```yaml
  ...
  mocks:
    service1:
      strategy: sequence
      sequence:
        # Отвечает разным текстом на каждый последующий запрос:
        # на первый запрос - "1", на второй - "2" и так далее.
        # Ответ на пятый и последующие запросы будет 404 Not Found.
        - strategy: constant
          body: '1'
        - strategy: constant
          body: '2'
        - strategy: constant
          body: '3'
        - strategy: constant
          body: '4'
    ...
```

##### basedOnRequest

Использует разные стратегии ответа, в зависимости от пути запрашиваемого ресурса. Разрешает объявлять несколько стратегий для одного пути. Конкурентно безопасен.

При получении запроса на ресурс, который не задан в параметрах, тест считается проваленным.

Параметры:

- `uris` (обязательный) - список ресурсов, каждый ресурс можно сконфигурировать как отдельный мок-сервис, используя любые доступные проверки запросов и стратегии ответов (см. пример)

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: basedOnRequest
      uris:
        - strategy: constant
          body: >
            {
              "ok": true
            }
          requestConstraints:
            - kind: queryMatches
              expectedQuery: "key=value1"
            - kind: pathMatches
              path: /request
        - strategy: constant
          body: >
            {
             "ok": true
            }
          requestConstraints:
            - kind: queryMatches
              expectedQuery: "key=value2"
            - kind: pathMatches
              path: /request
    ...
```

##### dropRequest

Стратегия, которая по умолчанию на любой запрос сбрасывает соединение. Используется для эмуляции проблем с сетью.

Не имеет параметров.

Пример:

```yaml
  ...
  mocks:
    service1:
      strategy: dropRequest
    ...
```

#### Подсчет количества вызовов

Вы можете указать, сколько раз должен быть вызван мок или отдельный ресурс мока (используя `uriVary`). Если фактическое количество вызовов будет отличаться от ожидаемого, тест будет считаться проваленным.

Пример:

```yaml
  ...
  mocks:
    service1:
      # должен вызываться ровно один раз
      calls: 1
      strategy: file
      filename: responses/books_list.json
  ...
```

```yaml
  ...
  mocks:
    service1:
      strategy: uriVary
      uris:
        /shelf/books:
          # должен вызываться ровно один раз
          calls: 1
          strategy: file
          filename: responses/books_list.json
  ...
```

## Использование shell скриптов

При запуске теста, операции выполняются в следующем порядке:

1. Загрузка фикстур
2. Запуск моков
3. Выполнение beforeScript
4. HTTP-запрос
5. afterRequestScript
6. Выполнение проверок

### Описание скрипта

Для описания скрипта нужно указать два параметра:

- `path` (обязательный) - строка, указывает путь к файлу скрипта.
- `timeout` - время в секундах, отвечает за завершение скрипта по таймауту. По-умолчанию таймаут будет равен `3`.

Пример:

```yaml
  ...
  afterRequestScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # таймаут будет равен 10с
    timeout: 10
  ...
```

```yaml
  ...
  beforeScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # таймаут будет равен 10с
    timeout: 10
  ...
```

```yaml
  ...
  beforeScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # таймаут будет равен 3с
  ...
```

### Запуск скрипта с параметризацией

В случае когда тесты используют параметризированные запросы также можно использовать различные скрипты для каждого запуска теста.

Пример:

```yaml
  ...
  beforeScript:
    path: |
      ./cli_scripts/{{.file_name}}
  ...
  cases:
    - requestArgs:
        customer_id: 1
        customer_email: "customer_1_recalculate@lamoda.ru"
      responseArgs:
        200:
          rrr: 1
          in_transit: 1
      beforeScriptArgs:
        file_name: "cmd_recalculate_customer_1.sh"
```

## Запрос в Базу данных

После выполнения http запросов можно выполнить SQL запрос в БД для проверки изменений данных.
Допускается что ответ может содержать несколько записей. Далее эти данные сравниваются с ожидаемым списком записей.

### Формат описания запросов

Для описания запросов к БД в тесте, можно использовать legacy-формат:

```yaml
- name: my test
  ...
  dbQuery: >
    SELECT ...
  dbResponse:
    - ...
    - ...
```

Но, более предпочтительным будет следующий формат:

```yaml
- name: my test
  ...
  dbChecks:
    - dbQuery: >
        SELECT ...
      dbResponse:
        - ...
```

Так как он, дает возможность выполнять более 1-го запроса в БД.
*ВАЖНО*: Все описанные ниже техники шаблонизации запросов, работают в обоих случаях описания теста.

### Описание запроса

Под запросом подразумевается SELECT, который вернет любое количество строк.

- `dbQuery` - строка, содержит SQL запрос

Пример:

```yaml
  ...
  dbQuery:
    SELECT code, purchase_date, partner_id FROM mark_paid_schedule AS m WHERE m.code = 'GIFT100000-000002'
  ...
```

### Описание ответа на запрос в Базу данных

Под ответом подразумевается список json объектов которые должен вернуть запрос в БД.

- `dbResponse` - строка, содержит список json объектов

Пример:

```yaml
  ...
  dbResponse:
    - '{"code":"GIFT100000-000002","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
    - '{"code":"GIFT100000-000003","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
    - '{"code":"$matchRegexp(GIFT([0-9]{6})-([0-9]{6}))","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
```

Как видно из примера, вы можете использовать Regexp для проверки ответа БД.

```yaml
  ...
  dbResponse:
    # пустой список
```

### Параметризация при запросах в Базу данных

Как и в случае с телом http-запроса, мы можем использовать параметризированные запросы.

Пример:

```yaml
  ...
    dbQuery: >
      SELECT code, partner_id FROM mark_paid_schedule AS m WHERE DATE(m.purchase_date) BETWEEN '{{ .fromDate }}' AND '{{ .toDate }}'

    dbResponse:
      - '{"code":"{{ .cert1 }}","partner_id":1}'
      - '{"code":"{{ .cert2 }}","partner_id":1}'
  ...
    cases:
      ...
      dbQueryArgs:
        fromDate: "2330-02-01"
        toDate: "2330-02-05"
      dbResponseArgs:
        cert1: "GIFT100000-000002"
        cert2: "GIFT100000-000003"
```

В случае, когда в разных тестах ответ содержит разное количество записей, вы можете переопределить ответ целиком для конкретного теста
продолжая использовать шаблон с параметрами в остальных

Пример:

```yaml
  ...
    dbQuery: >
      SELECT code, partner_id FROM mark_paid_schedule AS m WHERE DATE(m.purchase_date) BETWEEN '{{ .fromDate }}' AND '{{ .toDate }}'

    dbResponse:
      - '{"code":"{{ .cert1 }}","partner_id":1}'
  ...
    cases:
      ...
      dbQueryArgs:
        fromDate: "2330-02-01"
        toDate: "2330-02-05"
      dbResponseArgs:
        cert1: "GIFT100000-000002"
      ...
      dbQueryArgs:
        fromDate: "2330-02-01"
        toDate: "2330-02-05"
      dbResponseFull:
        - '{"code":"GIFT100000-000002","partner_id":1}'
        - '{"code":"GIFT100000-000003","partner_id":1}'
```

### Игнорирование порядка записей в ответе на запрос в базу данных

Можно использовать флаг `ignoreDbOrdering` в секции `comparisonParams` для включения/выключения функционала проверки полученных строк в ответе от базы данных не по порядку.
Это может пригодиться для обхода использования оператора `ORDER BY` в запросах.

- `ignoreDbOrdering` - значение true/false.

Пример:

```yaml
  comparisonParams:
    ignoreDbOrdering: true
  ...
  dbQuery: >
    SELECT id, name, surname FROM users LIMIT 2
    
  dbResponse:
    - '{ "id": 2, "name": "John", "surname": "Doe" }'
    - '{ "id": 1, "name": "Jane", "surname": "Doe" }'
```

## JSON-schema

Для упрощения написания тестов на Gonkey, используйте [файл со схемой](https://raw.githubusercontent.com/lamoda/gonkey/master/gonkey.json)

Он добавляет in-line документацию и авто-дополнение в IDE которые это поддерживают.

Пример работы в IDE Jetbrains:
![Example Jetbrains](https://i.imgur.com/oYuPuR3.gif)

Пример работы в IDE VSCode:
![Example Jetbrains](https://i.imgur.com/hBIGjP9.gif)

### Настройка на IDE Jetbrains

Скачайте [файл со схемой](https://raw.githubusercontent.com/lamoda/gonkey/master/gonkey.json).
В настройках Languages & Frameworks > Schemas and DTDs > JSON Schema Mappings

![Jetbrains IDE Settings](https://i.imgur.com/xkO22by.png)

Добавьте новую схему

![Add schema](https://i.imgur.com/XHw14GJ.png)

Задайте имя схемы, выберите скачанный файл, а также выберите версию схем Draft 7

![Name, file, version](https://i.imgur.com/LfJfis0.png)

После этого добавьте маппинг. Можно выбрать файл, папку с тестами или маску.

![Mapping](https://i.imgur.com/iFjm0Ld.png)

Выберите то что удобно для вас.

![Mapping pattern](https://i.imgur.com/WIK6sZW.png)

После этого сохраните настройки, и если вы всё сделали правильно в нижнем правом углу окна IDE не должно
отображатся No JSON Schema
![No Schema](https://i.imgur.com/zLqv1Zv.png)

А должно быть ваше название схемы.

![Schema Name](https://i.imgur.com/DDXdCO7.png)

### Настройка на IDE VSCode

Для начала вам нужно установить плагин для работы с YAML
Откройте меню Code(File)->Preferences->Extensions

![VSCode Preferences](https://i.imgur.com/X7bk5Kh.png)

Наберите в поиске YAML, и установите расширение YAML Language Support by Red Hat

![Yaml Extension](https://i.imgur.com/57onioF.png)

Откройте меню Code(File)->Preferences->Settings
Наберите YAML:Schemas и нажмите на ссылку *Edit in settings.json*
![Yaml link](https://i.imgur.com/IEwxWyG.png)

Добавьте маппинг файла и путь к схеме

```
"yaml.schemas": {
  "C:\\Users\\Leo\\gonkey.json": ["*.gonkey.yaml"]          
}
```

В примере выше, схема из файла C:\Users\Leo\gonkey.json будет применяться ко всем файлам
с расширением .gonkey.yaml

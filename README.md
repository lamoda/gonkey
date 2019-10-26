# Gonkey: testing automation tool

[View Russian version of this file here](README-ru.md)

Gonkey will test your services using their API. It can bomb the service with prepared requests and check the responses. Test scenarios are described in YAML-files.

Capabilities:
- works with REST/JSON API
- tests service API for compliance with OpenAPI-specs
- seeds the DB with fixtures data (supports PostgreSQL)
- provides mocks for external services
- can be used as a library and ran together with unit-tests
- stores the results as an [Allure](http://allure.qatools.ru/) report

### Using the CLI

To test a service located on a remote host, use gonkey as a console util.

`./gonkey -host <...> -tests <...> [-spec <...>] [-db_dsn <...> -fixtures <...>] [-allure] [-v]`

- `-spec <...>` path to a file or URL with the swagger-specs for the service
- `-host <...>` service host:port
- `-tests <...>` test file or directory
- `-db_dsn <...>` DSN for the test DB (the DB will be cleared before seeding!), supports only PostgreSQL
- `-fixtures <...>` fixtures directory
- `-allure` generate an Allure-report
- `-v` verbose output
- `-debug` debug output

You can't use mocks in this mode.

### Using gonkey as a library

To integrate functional and native Go tests and run them together, use gonkey as a library.

Create a test file, for example `func_test.go`.

Import gonkey as a dependency to your project in this file.

```go
import (
	"github.com/lamoda/gonkey/runner"
	"github.com/lamoda/gonkey/mocks"
)
```

Create a test function.

```go
func TestFuncCases(t *testing.T) {
    // init the mocks if needed (details below)
    //m := mocks.NewNop(...)

    // init the DB to load the fixtures if needed (details below)
    //db := ...

    // create a server instance of your app
    srv := server.NewServer()
    defer srv.Close()

    // run test cases from your dir with Allure report generation
    runner.RunWithTesting(t, &runner.RunWithTestingParams{
        Server:      srv,
        TestsDir:    "cases",
        Mocks:       m,
        DB:          db,
        FixturesDir: "fixtures",
    })
}
```

The tests can be now ran with `go test`, for example: `go test ./...`.

### Test file example
```yaml
- name: WHEN the list of orders is requested MUST successfully response
  method: GET
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

- name: WHEN one order is requested MUST response with user and sum
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

### HTTP-request

`method` - a parameter for HTTP request type, the format is in the example above.

`path` - a parameter for URL path, the format is in the example above.

`headers` - a parameter for HTTP headers, the format is in the example above.

`cookies` - a parameter for cookies, the format is in the example above.


### Fixtures

To seed the DB before the test, gonkey uses fixture files.

File example:

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

Records in fixtures can use templates, inherit and reference each other.

#### Record templates

Usually, to insert a record to a DB, it's necessary to list all the fields without default values. Oftentimes, many of those fields are not important for the test and their values repeat from one fixture to another, creating unnecessary visual garbage and making the maintenance harder.

With templates you can inherit the fields from template record redefining only the fields that are important for the test.

Template definition example:
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

Example of using a template in a fixture:
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

As you might have noticed, templates can be inherited as well with `$extend` keyword, but only if by the time of the dependent template definition the parent template is already defined (in this file or any other referended with `inherits`).

#### Record inheritance

Records can be inherited as well using `$extend`.

To inherit a record, first you need to assign this record a name using `$name`:

```yaml
# fixtures/post.yaml
tables:
  posts:
    - $name: regular_post
      title: Post title
      text: Some text
```

Names assigned to records must be unique among all loaded fixture files, as well as they must not interfere with template names.

In another fixture file you need to declare that a certain record inherits an earlier defined record with `$extend`, just like with the templates:

```yaml
# fixtures/deleted_post.yaml
inherits:
  - post
tables:
  posts:
    - $extend: regular_post
      is_deleted: true
```

Don't forget to declare the dependency between files in `inherits`, to make sure that one file is always loaded together with the other one.

It's important to note that record inheritance only works with different fixture files. It's not possible to declare inheritance within one file.

#### Record linking

Despite the fact that fixture files allow you to set values for autoincrement columns (usually `id`), it's not recommended to do it. It's very difficult to control that all the values for `id` are correct between different files and that they never interfere. In order to let the DB assign autoincrement values it's enough to not set the value explicitly.

However, if the value for `id` is not set explicitly, how is it possible to link several entities that should reference each other with ids? Fixtures let us to reference previously inserted records by their name, using `$refName.fieldName`.

Let's declare a named record:

```yaml
# fixtures/post.yaml
tables:
  posts:
    - $name: regular_post
      title: Post title
      text: Some text
```

Now, in order to link `posts` and `comments` tables, we can address the record using its name (`$regular_post`) and pass the field where the value should be taken from (`id`):

```yaml
# fixtures/comment.yaml
tables:
  comments:
    - post_id: $regular_post.id
      content: A comment...
      author_name: John Doe
```

You can only reference fields of a previously inserted record. It's impossible to reference template fields, when trying to do that you'll get an `undefined reference` error.

Take a note of a limitation: you can't reference records within one table of one file.

#### Expressions

When you need to write an expression execution result to the DB and not a static value, you can use `$eval()` construct. Everything inside the brackets will be inserted into the DB as raw, non-escaped data. This way, within `$eval()` you can write everything you would in a regular query.

For instance, this construct allows the insertion of current date and time as a field value:

```yaml
tables:
  comments:
    - created_at: $eval(NOW())
```

### Mocks

In order to imitate responses from external services, use mocks.

A mock is a web server that is running on-the-fly, and is populated with certain logic before the execution of each test. The logic defines what the server responses to a certain request. It's defined in the test file.

#### Running mocks while using gonkey as a library

Before running tests, all planned mocks are started. It means that gonkey spins up the given number of servers and each one of them gets a random port assigned.

```go
// create empty server mocks
m := mocks.NewNop(
	"cart",
	"loyalty",
	"catalog",
	"madmin",
	"okz",
	"discounts",
)

// spin up mocks
err := m.Start()
if err != nil {
    t.Fatal(err)
}
defer m.Shutdown()
```

After spinning up the mock web-servers, we can get their addresses (host and port). Using those addresses, you can configure your service to send their requests to mocked servers instead of real ones.

```go
// configuring and running the service
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

As soon as you spinned up your mocks and configured your service, you can run the tests.

```go
runner.RunWithTesting(t, &runner.RunWithTestingParams{
    Server:    srv,
    Directory: "tests/cases",
    Mocks:     m, // passing the mocks to the test runner
})
```

#### Mocks definition in the test file

Each test communicates a configuration to the mock-server before running. This configuration defines the responses for specific requests in the mock-server. The configuration is defined in a YAML-file with test in the `mocks` section.

The test file can contain any number of mock service definitions:

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

Each mock-service definition consists of:

`requestConstraints` - an array of constraints that are applied on a received request. If at least one constraint is not satisfied, the test is considered failed. The list of all possible checks is provided below.

`strategy` - the strategy of mock responses. The list of all possible strategies is provided below.

The rest of the keys on the first nesting level are parameters to the strategy. Their variety is different for each strategy.

A configuration example for one mock-service:
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

##### Request constraints (requestConstraints)

The request to the mock-service can be validated using one or more constraints defined below.

The definition of each constraint contains of the `kind` parameter that indicates which constraint will be applied.

All other keys on this level are constraint parameters. Each constraint has its own parameter set.

###### nop

Empty constraint. Always successful.

No parameters.

Example:
```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: nop
    ...
```

###### bodyMatchesJSON

Checks that the request body is JSON and it corresponds to the JSON defined in the `body` parameter.

Parameters:
- `body` (mandatory) - expected JSON. All keys on all levels defined in this parameter must be present in the request body.

Example:
```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        # this check will demand that the request contains keys key1, key2 and subKey1
        # and their values set to value1 and value2. However, it's fine if the request has 
        # other keys not mentioned here.
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

###### expectedQuery

Checks that the GET request parameters correspond to the ones defined in the `query` paramter.

Parameters:
- `query` (mandatory) - a list of parameters to compare the parameter string to. The order of parameters is not important.

Example:
```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        # this check will demand that the request contains key1 и key2
        # and the values are key1=value1, key1=value11 и key2=value2. 
        # Keys not mentioned here are omitted while running the check.
        - kind: expectedQuery
          query:  key1=value1&key2=value2&key1=value11
    ...
```

###### methodIs

Checks that the request method corresponds to the expected one.

Parameters:
- `method` (mandatory) - string to compare the request method to.

There are also 2 short variations that don't require `method` parameter:
- `methodIsGET`
- `methodIsPOST`

Examples:
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

###### headerIs

Checks that the request has the defined header and (optional) that its value either equals the pre-defined one or falls under the definition of a regular expression.

Parameters:
- `header` (mandatory) - name of the header that is expected with the request;
- `value` - a string with the expected request header value;
- `regexp` - a regular expression to check the header value against.

Examples:
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

##### Response strategies (strategy)

Response strategies define what mock will response to incoming requests.

###### nop

Empty strategy. All requests are served with `204 No Content` and empty body.

No parameters.

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: nop
    ...
```

###### file

Returns a response read from a file.

Parameters:
- `filename` (mandatory) - name of the file that contains the response body;
- `statusCode` - HTTP-code of the response, the default value is `200`;
- `headers` - response headers.

Example:
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

###### constant

Returns a defined response.

Parameters:
- `body` (mandatory) - sets the response body;
- `statusCode` - HTTP-code of the response, the default value is `200`;
- `headers` - response headers.

Example:
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

###### uriVary

Uses different response strategies, depending on a path of a requested resource.

When receiving a request for a resource that is not defined in the parameters, responses with `404 Not Found`.

Parameters:
- `uris` (mandatory) - a list of resources, each resource can be configured as a separate mock-service using any available request constraints and response strategies (see example)
- `basePath` - common base route for all resources, empty by default

Example:
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

###### methodVary

Uses various response strategies, depending on the request method.

When receiving a request with a method not defined in methodVary, the server responses with `405 Method Not Allowed`.

Parameters:
- `methods` (mandatory) - a list of methods, each method can be configured as a separate mock-service using any available request constraints and response strategies (see example)

Example:
```yaml
  ...
  mocks:
    service1:
      strategy: methodVary
      methods:
        GET:
          # nothing stops us from using `uriVary` strategy here
          # this way we can form different responses to different 
          # method+resource combinations
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

##### Calls count

You can define, how many times each mock or mock resource must be called (using `uriVary`). If the actual number of calls is different from expected, the test will be considered failed.

Example:
```yaml
  ...
  mocks:
    service1:
      # must be called exactly one time
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
          # must be called exactly one time
          calls: 1
          strategy: file
          filename: responses/books_list.json
  ...
```

### CMD interface

Before running an HTTP request you can run a script using cmd interface.
When the test is ran, the first step is to load fixtures and run mocks. Next, the script is executed, then the test is ran.

#### Script definition

To define the script you need to provide 2 parameters:

- `path` (mandatory) - string with a path to the script file.
- `timeout` - time in seconds, is responsible for stopping the script on timeout. The default value is `3`.

Example:
```yaml
  ...
  beforeScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # the timeout will equal 10s
    timeout: 10
  ...
```

```yaml
  ...
  beforeScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # the timeout will equal 3s
  ...
```

#### Running a script with parameterization

When tests use parameterized requests, it's possible to use different scripts for each test run.

Example:
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
        scriptArgs:
          file_name: "cmd_recalculate_customer_1.sh"
```

### A DB query

After HTTP request execution you can run an SQL query to DB to check the data changes.
The response can contain several records. Those records are compared to the expected list of records.

#### Query definition

Query is a SELECT that returns any number of strings.

- `dbQuery` - a string that contains an SQL query.

Example:
```yaml
  ...
  dbQuery:
    SELECT code, purchase_date, partner_id FROM mark_paid_schedule AS m WHERE m.code = 'GIFT100000-000002'
  ...
```

#### Definition of DB request response

The response is a list of JSON bojects that the DB request should return.

- `dbResponse` - a string that contains a list of JSON objects.

Example:
```yaml
  ...
  dbResponse:
    - '{"code":"GIFT100000-000002","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
    - '{"code":"GIFT100000-000003","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
```

```yaml
  ...
  dbResponse:
    # empty list
```
#### DB request parameterization

As well as with the HTTP request body, we can use parameterized requests.

Example:
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

When different tests contain different number of records, you can redefine the response for a specific test as a whole, while continuing to use a template with parameters in others.

Example:
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


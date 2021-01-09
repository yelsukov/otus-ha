### Content

* [Description](#Description)
    * [Functional requirements](#Functional-requirements)
    * [Non functional requirements](#Non-functional-requirements)
    * [Security requirements](#Security-requirements)
* [How to run](#How-to-run)
* [Docker images](#Docker-images)
* [Technology stack](#Technology-stack)
* [Makefile commands](#Makefile-commands)
    * [List of parameters](#List-of-parameters)
    * [List of actions](#List-of-actions)
* [Project structure](#Project-structure)

### Description

Task: Create template of the social network for HL-tests and fun.

#### Functional requirements

- Authorization by username and password.
- Registration page with next data:
    1. Name
    2. Surname
    3. Age
    4. Gender
    5. Interests
    6. City
- User profile pages
- Friends list
- Possibility to make/break a friendship
- Create and update a profile

#### Non functional requirements

- Any programming language
- MySQL as Database
- Do not use ORM
- Monolith architecture
- Do not use next technologies:
    1) Replication
    2) Sharding
    3) Indexes
    4) Caching
- The html layout is not important. The most primitive will be ok.

#### Security requirements

- Secured SQL-injections
- Safe storage of passwords


### How to run

> You need docker and docker-compose installed.

> Ports 3336 (MySQL), 8007 (Backend) and 8008 (Frontend) should be available

Just execute in terminal `make up`

The command will build the Docker images and runs the containerы using a compose file.  

Or you can run docker-compose directly:
`sudo docker-compose up --build -d`

You will find the frontend SPA at http://127.0.0.1:8008

If you want to try direct requests to API, then you can find it at http://127.0.0.1:8007/v1

### Docker images

List of used images

|Name|Base Image|Size|Description|
|-----|------|------|------|
|`mysqldb`|mysql:8.0.22|**545MB**|Plain image with mysql db|
|`backend`|golang:1.15-alpine -> scratch|**299MB + 8.35MB**|Backend server with API|
|`frontend`|node:14.15-alpine -> nginx:1.19.6-alpine|**116MB + 22.3MB**|Frontend server based on nginx|

### Technology stack

|Service|Technologies|Description|
|-----|------|------|
|`Database`|MySQL v8.0.22|RDBS|
|`Backend`|golang v1.15|Restful API service|
|`frontend`|Angular v10, NodeJs v14, Typescript, nginx v1.19|SPA with http server|


### Makefile commands

You can do some of the routine work with simple commands from the Makefile.

To use it just run the following console command from your project folder:
> $ make [action]

#### List of parameters

|Name|Description|Default|Possible values|
|---|---|---|---|
|GOOS|Desired type of platform for building binaries|linux|darwin, dragonfly, freebsd, linux, nacl, netbsd, openbsd, plan9, solaris, windows|
|GOARCH|Desired type of architecture for building binaries|amd64|386, amd64, amd64p32, arm, arm64, ppc64, ppc64le, mips, mipsle, mips64, mips64le, s390x|
|CGO_ENABLED|Enable or disable cgo tool|0|`0` or `1`|

#### List of actions

|Action|Description|
|---|---|
|`up`|Creates & runs all needed containers|
|`down`|Stops docker network and removes all containers|
|`buildBackend`|Compiles binaries for backend service|
|`buildFrontend`|Builds SPA for frontend|
|`fmt`|Runs goimports for all *.go files in project|
|`clean`|Removes all compiled binaries and static files of SPA|


### Project structure

```

├── backend                  # Backend API server
│   ├── config                 # Configuration package
│   ├── errors                 # Custom errors package
│   ├── jwt                    # JWT implementation package
│   ├── migrations             # List of mirgations for DB
│   ├── models                 # Models description
│   ├── server                 # Implementation of web server
│   │    └── v1                # API v1
│   │        ├── handlers        # Handlers of API actions
│   │        ├── requests        # Structures of requests 
│   │        └── responses       # Structures of responses
│   └── storages               # Packages for interaction with DB
└── frontend                 # Fronentd SPA
    └── src
        ├── app                # SPA core
        │   ├── components       # List of components and controllers
        │   ├── helpers          # List of helpers
        │   ├── interceptors     # Interceptors for Auth and Erros
        │   ├── models           # Models classes
        │   ├── services         # Services and providers
        │   └── validators       # Custom validators
        └── environments         # Env configuration
```

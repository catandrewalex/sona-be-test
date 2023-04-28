# Sonamusica Administration Backend

![GitHub repo size](https://img.shields.io/github/repo-size/FerdiantJoshua/sonamusica-administration-backend) ![GitHub issues](https://img.shields.io/github/issues/FerdiantJoshua/sonamusica-administration-backend) ![GitHub](https://img.shields.io/github/license/FerdiantJoshua/sonamusica-administration-backend) ![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/FerdiantJoshua/sonamusica-administration-backend?include_prereleases)

Backend part for the administration app of Sonamusica, a private music school in Bandung. This app manages teacher payroll, student & teacher presences, and many more. 

## Requirement

- [Golang](https://go.dev/) 1.9 or higher (tested in 1.9)

## Execution

### Windows

```bat
.\scripts\build_and_run.bat
```

### Linux & MacOS

```sh
./scripts/build_and_run.sh
```

## Migrate Database

### Windows

```bat
.\scripts\mysql\reset_and_migrate_database.bat
```

### Linux or MacOS

```sh
./scripts/mysql/reset_and_migrate_database.sh
```

## Generate SQL Queries using SQLC

### Windows

#### MySQL

1. Download the binary files [here](https://github.com/kyleconroy/sqlc/releases/download/v1.17.2/sqlc_1.17.2_windows_amd64.zip)

2. Call SQLC's `generate`

    ```sh
    sqlc generate
    ```

#### PostgreSQL

1. Pull Docker image

    ```sh
    docker pull kjconroy/sqlc
    ```

2. Call SQLC's `generate` by executing the Docker image on our directory, then immediately remove the container

    ```sh
    docker run --rm -v "$(pwd):/src" -w /src kjconroy/sqlc generate
    ```

## Environment Variables

1. Rename `.env.example` to `.env`
2. Modify and adjust any variables according to your configuration

## Contributor

- Author: [FerdiantJoshua](https://github.com/FerdiantJoshua)

## License

[MIT](LICENSE)

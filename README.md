# device-analytics

Device Analytics is a service that provides a public API developed and maintained by the Wikimedia Foundation that serves the [Unique Devices dataset](https://wikitech.wikimedia.org/wiki/Analytics/AQS/Unique_Devices), which contains the number of unique devices that have visited a Wikimedia project over a given period of time.

### Docker Quickstart

You will need:
- [aqs-docker-test-env](https://gitlab.wikimedia.org/frankie/aqs-docker-test-env) and its associated dependencies

Start up the Dockerized test environment in aqs-docker-test-env and load 

```sh-session
make startup
```

then:

```sh-session
go run .
```
Then, connect to `http://localhost:8080/`.

## Unit Testing

To run a suite of unit tests, first start up the Dockerized test environment in aqs-docker-test-env, then:

```sh-session
make test
```

## API documentation

To generate and view the API documentation, read the [docs on Wikitech][wikipage].

[wikipage]: https://wikitech.wikimedia.org/wiki/AQS_2.0#API_documentation

# zipapi
A small coding challenge project.

<a href="https://travis-ci.org/romshark/zipapi">
	<img src="https://travis-ci.org/romshark/zipapi.svg?branch=master" alt="Travis CI: build status">
</a>
<a href='https://coveralls.io/github/romshark/zipapi'>
	<img src='https://coveralls.io/repos/github/romshark/zipapi/badge.svg' alt='Coverage Status' />
</a>
<a href="https://goreportcard.com/report/github.com/romshark/zipapi">
	<img src="https://goreportcard.com/badge/github.com/romshark/zipapi" alt="GoReportCard">
</a>

----


[zipapi](https://github.com/romshark/zipapi) is an HTTP(S) API that takes files uploaded to `POST /archive` as `multipart/form-data`
into a zip archive and returns it as a response.

## Roadmap

- Required:
	- [x] Upload of multiple files and creating a zip file out of them
	- [x] Provide the zip file back to the API user 
	- [x] Serve multiple requesting clients at the same time
	- [x] Limit the size of each uploaded file
- Nice to have:
	- [x] Provide automated API tests
	- [x] CI
	- [x] Retain a history of all created Zip files and their contents *
	- [ ] Expire created Zip files after a specific period of time **

_\* There's currently no database-backed persistency implementation but just a simple in-memory mock. Implementing one shouldn't be a problem though._

_\** This service shouldn't be responsible for this problem. There must be a separate service that periodically goes over the database and cleans up expired files._

## Getting started

- Download the latest release from [releases](https://github.com/romshark/zipapi/releases)
- Compile `/cmd/zipapi` using `go build`
- Define a configuration using `/cmd/zipapi/config.toml` as a template
- Run the server using `./zipapi -config /path/to/config.toml` providing the path to your configuration file.

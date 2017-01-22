## CRDTs over Kafka

This repo will contains sample implementations of the catalog of CRDTs at http://hal.upmc.fr/inria-00555588/document, implemented over Kafka.


To run the tests, bring up a local kafka environment:

```
docker-compose up -d
```

Get the test dependencies:
```
go get -t ./...
```

Run the tests:
```
go test ./...
```

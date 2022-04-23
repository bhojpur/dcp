# Bhojpur DCP - Statebase

The `statebase` framework is an `etcd` shim that translates etcd API to SQLite,
PostgreSQL, MySQL, and dqlite.

## Features

- Can be run standalone so any Kubernetes (not just Bhojpur DCP) can use it
- Implements a subset of etcdAPI (not usable at all for general purpose etcd)
- Translates etcdTX calls into the desired API (Create, Update, Delete)
- Backend drivers for dqlite, sqlite, Postgres, MySQL and NATS JetStream

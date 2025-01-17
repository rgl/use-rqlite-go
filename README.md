# About

[![Build](https://github.com/rgl/use-rqlite-go/actions/workflows/main.yml/badge.svg)](https://github.com/rgl/use-rqlite-go/actions/workflows/main.yml)

My [rqlite](https://github.com/rqlite/rqlite) Go example.

# Usage

This can be tested [in docker compose](#usage-docker-compose) or [in a kind kubernetes cluster](#usage-kubernetes).

List this repository dependencies (and which have newer versions):

```bash
GITHUB_COM_TOKEN='YOUR_GITHUB_PERSONAL_TOKEN' ./renovate.sh
```

## Usage (docker compose)

Install docker and docker compose.

Create the environment:

```bash
docker compose up --build --detach
docker compose ps
docker compose logs
```

Start the rqlite client:

```bash
docker compose exec rqlite rqlite
```

Execute some commands:

```
.status
.nodes
.schema
select sqlite_version()
insert into quote(id, author, text, url) values(100, "Hamlet", "To be, or not to be, that is the question.", null)
select text || ' -- ' || author as quote from quote
.exit
```

Get a quote from the service:

```bash
wget -qO- http://localhost:4000
```

Destroy the environment:

```bash
docker compose down --volumes --remove-orphans
```

## Usage (Kubernetes)

Install docker, kind, kubectl, and helm.

Create the local test infrastructure:

```bash
./.github/workflows/kind/create.sh
```

Access the test infrastructure kind Kubernetes cluster:

```bash
export KUBECONFIG="$PWD/kubeconfig.yml"
kubectl get nodes -o wide
```

Build and use the use-rqlite-go example:

```bash
./build.sh && ./deploy.sh && ./test.sh && xdg-open index.html
```

Start the rqlite client:

```bash
kubectl exec --quiet --stdin --tty statefulset/rqlite -- rqlite
```

Execute some commands:

```
.status
.nodes
.schema
select sqlite_version()
insert into quote(id, author, text, url) values(100, "Hamlet", "To be, or not to be, that is the question.", null)
select text || ' -- ' || author as quote from quote
.exit
```

Destroy the local test infrastructure:

```bash
./.github/workflows/kind/destroy.sh
```

# References

* [rqlite homepage](https://rqlite.io)
* [rqlite source-code repositories](https://github.com/rqlite)
* [rqlite Go client (gorqlite)](https://github.com/rqlite/gorqlite)

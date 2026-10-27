# Go on Master da Web

A small, fast HTTP service in Go, with private files on Object Storage.

| Service | What it runs |
| --- | --- |
| `api` | the binary `go build` makes |

The private bucket comes as `MW_STORAGE_*` and is used with the AWS SDK for Go v2; each download link is signed and expires in 15 minutes.

`POST /files` stores a file; `GET /files` lists them with their links. The home page does the same from the browser.

## Make it yours

1. In the console, open the project and follow **Your code**: copy this
   template to your GitHub and connect the copy to the project.
2. Change `main.go`, then push: every push deploys, with no downtime.

## Run it on your machine

```bash
export $(grep -v ^# .env.example | xargs)
go run .
```

## Good to know

- The service listens on `0.0.0.0:$PORT`: the platform sets `PORT`.

*[Português](README.md)*

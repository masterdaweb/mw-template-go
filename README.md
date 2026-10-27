# Go na Master da Web

Um serviço HTTP em Go, pequeno e rápido, com arquivos privados no Object Storage.

| Serviço | O que roda |
| --- | --- |
| `api` | o binário que o `go build` gera |

O bucket privado chega como `MW_STORAGE_*` e é usado com o AWS SDK for Go v2; cada link de download é assinado e expira em 15 minutos.

`POST /files` guarda um arquivo; `GET /files` lista com os links. A página inicial faz o mesmo pelo navegador.

## Torne o código seu

1. No console, abra o projeto e siga **Seu código**: copie este template
   para o seu GitHub e conecte a cópia ao projeto.
2. Altere `main.go` e faça push: cada push faz um deploy, sem tirar o
   app do ar.

## Rodar na sua máquina

```bash
export $(grep -v ^# .env.example | xargs)
go run .
```

## Bom saber

- O serviço escuta em `0.0.0.0:$PORT`: a plataforma define `PORT`.

*[English](README.en.md)*

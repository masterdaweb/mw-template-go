package main

import "strings"

// texts: the page in the visitor's language, Portuguese when the browser prefers it.
func texts(acceptLanguage string) (string, map[string]string) {
	first, _, _ := strings.Cut(acceptLanguage, ",")
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(first)), "pt") {
		return "pt-BR", pt
	}
	return "en", en
}

var en = map[string]string{
	"title":    "Your Go service is live",
	"lead":     "A small, fast HTTP service with private files on Object Storage: files sent below go to the bucket and come back through links that expire in 15 minutes.",
	"reaches":  "What this service reaches",
	"send":     "Send a file",
	"file":     "File",
	"submit":   "Send to Object Storage",
	"sent":     "%s was saved.",
	"empty":    "Choose a file first.",
	"failed":   "The file could not be saved. See the service logs.",
	"none":     "No files yet.",
	"download": "download (link valid for 15 min)",
	"api":      "The same as JSON",
	"next":     "Next steps",
	"step1":    "Make the code yours: in the console, open this project and follow “Your code” to copy it to your GitHub and connect it.",
	"step2":    "Change main.go, then push: every push builds and deploys, with no downtime.",
	"step3":    "Add your own domain on the api service, under Domains.",
}

var pt = map[string]string{
	"title":    "Seu serviço Go está no ar",
	"lead":     "Um serviço HTTP pequeno e rápido, com arquivos privados no Object Storage: os arquivos enviados abaixo vão para o bucket e voltam por links que expiram em 15 minutos.",
	"reaches":  "O que este serviço alcança",
	"send":     "Envie um arquivo",
	"file":     "Arquivo",
	"submit":   "Enviar para o Object Storage",
	"sent":     "%s foi salvo.",
	"empty":    "Escolha um arquivo primeiro.",
	"failed":   "Não foi possível salvar o arquivo. Veja os logs do serviço.",
	"none":     "Nenhum arquivo ainda.",
	"download": "baixar (link válido por 15 min)",
	"api":      "O mesmo em JSON",
	"next":     "Próximos passos",
	"step1":    "Torne o código seu: no console, abra este projeto e siga “Seu código” para copiá-lo para o seu GitHub e conectá-lo.",
	"step2":    "Altere main.go e faça push: cada push faz o build e o deploy, sem tirar o serviço do ar.",
	"step3":    "Adicione o seu domínio no serviço api, em Domínios.",
}

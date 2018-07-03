package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/howeyc/gopass"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/net/publicsuffix"
)

type Client struct {
	httpClient http.Client
}

func Login(login string, password string) Client {
	_url := "https://sistemas.uepg.br/academicoonline/login/index"
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	httpClient := http.Client{Jar: jar}
	aux, _ := httpClient.Get(_url)
	jsession := aux.Header["Set-Cookie"][0][:43]

	form := url.Values{}
	form.Set("login", login)
	form.Set("password", password)

	req, _ := http.NewRequest(
		"POST",
		"https://sistemas.uepg.br/academicoonline/login/authenticate",
		bytes.NewBufferString(form.Encode()))

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("cookie", jsession)

	resp, _ := httpClient.Do(req)

	defer resp.Body.Close()

	client := Client{httpClient}
	return client
}

func (client Client) Grade() ([]string, [][]string) {
	_url := "https://sistemas.uepg.br/academicoonline/avaliacaoDesempenho/index"

	resp, _ := client.httpClient.Get(_url)
	defer resp.Body.Close()

	var contents []string

	doc, _ := goquery.NewDocumentFromResponse(resp)
	doc.Find("td").Each(func(i int, content *goquery.Selection) {
		contents = append(contents, content.Text())
	})

	header := []string{"Código", "Nome", "Turma", "Calendário",
		"Faltas", "Nota1", "Nota2", "NotaEx", "Média", "Freq", "Situação"}

	var table [][]string
	for i := 0; i < len(contents); i += 11 {
		table = append(table, contents[i:i+11])
	}

	return header, table
}

func main() {
	var login string

	fmt.Printf("Login: ")
	fmt.Scanf("%s", &login)

	fmt.Printf("Password: ")
	password, _ := gopass.GetPasswdMasked()

	client := Login(login, string(password))

	fmt.Println(`
  __   __  _______  _______  _______  _______  _______  _______  ______   _______  __    _  ___      ___   __    _  _______
 |  | |  ||       ||       ||       ||   _   ||       ||   _   ||      | |       ||  |  | ||   |    |   | |  |  | ||       |
 |  | |  ||    ___||    _  ||    ___||  |_|  ||       ||  |_|  ||  _    ||   _   ||   |_| ||   |    |   | |   |_| ||    ___|
 |  |_|  ||   |___ |   |_| ||   | __ |       ||       ||       || | |   ||  | |  ||       ||   |    |   | |       ||   |___
 |       ||    ___||    ___||   ||  ||       ||      _||       || |_|   ||  |_|  ||  _    ||   |___ |   | |  _    ||    ___|
 |       ||   |___ |   |    |   |_| ||   _   ||     |_ |   _   ||       ||       || | |   ||       ||   | | | |   ||   |___
 |_______||_______||___|    |_______||__| |__||_______||__| |__||______| |_______||_|  |__||_______||___| |_|  |__||_______|
    `)

	header, grade := client.Grade()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(grade)

	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

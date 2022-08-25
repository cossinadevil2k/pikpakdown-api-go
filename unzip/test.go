package main

import (
	"fmt"
	"github.com/mrxtryagin/pikpakdown-api-go/myzip"
	"strconv"
	"time"
)

func main() {
	u1 := "https://lgte-my.sharepoint.com/personal/mrx_lostknife_win/_layouts/15/download.aspx?UniqueId=ea306b0b-cc21-4b8b-a3cb-5f17992fd0cb&Translate=false&tempauth=eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJhdWQiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAvbGd0ZS1teS5zaGFyZXBvaW50LmNvbUA2Nzg0NTYxYS1lNTJkLTRlZGUtYmY4Yy1lZjBmNjk0ZGU5ZjIiLCJpc3MiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAiLCJuYmYiOiIxNjYxMzM1NzM4IiwiZXhwIjoiMTY2MTMzOTMzOCIsImVuZHBvaW50dXJsIjoiRjJUZ05OdTNKOG01R1YxQVlGbTIyQzdvaEZuaXI0UTBRQXhsdTlOMjlJVT0iLCJlbmRwb2ludHVybExlbmd0aCI6IjE0NSIsImlzbG9vcGJhY2siOiJUcnVlIiwiY2lkIjoiWlRNM1l6QXlaV1F0TXpGaE55MDBaVGMzTFdJNU1UVXRPVFF6TVRReU1XSXhaalZsIiwidmVyIjoiaGFzaGVkcHJvb2Z0b2tlbiIsInNpdGVpZCI6IlpUUmxZamt4TlRFdE0yUTVNeTAwWXpNM0xXRmhZVEV0TlRBeFpUTmtNMlpoTjJKayIsImFwcF9kaXNwbGF5bmFtZSI6ImNsb3VkcmV2ZSIsImdpdmVuX25hbWUiOiJyeCIsImZhbWlseV9uYW1lIjoibSIsImFwcGlkIjoiNjQ0NWJkNWItZjI1OS00YmY2LTgxMTItZGFjODA2N2RmZjM5IiwidGlkIjoiNjc4NDU2MWEtZTUyZC00ZWRlLWJmOGMtZWYwZjY5NGRlOWYyIiwidXBuIjoibXJ4QGxvc3RrbmlmZS53aW4iLCJwdWlkIjoiMTAwMzIwMDBBNzQ4Q0IzOCIsImNhY2hla2V5IjoiMGguZnxtZW1iZXJzaGlwfDEwMDMyMDAwYTc0OGNiMzhAbGl2ZS5jb20iLCJzY3AiOiJhbGxmaWxlcy53cml0ZSIsInR0IjoiMiIsInVzZVBlcnNpc3RlbnRDb29raWUiOm51bGwsImlwYWRkciI6IjIwLjE5MC4xNDQuMTcxIn0.a25HZUpCWmhKYzBOc3l2ZGxVVjYxdXBQV2w0R1Y0OTBFeVFCL1Bjakl5bz0&ApiVersion=2.0"
	reader, err := myzip.GetZipReaderFromUrl(u1)
	if err != nil {
		panic(err)
	}
	myzip.PrintZipFiles(reader)
	folder := fmt.Sprintf("%s", strconv.Itoa(int(time.Now().Unix())))
	_, err = myzip.UnZipFilesFromNumbers(u1, reader, folder, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	if err != nil {
		panic(err)
	}
}

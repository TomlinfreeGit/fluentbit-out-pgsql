package utils

import (
	"fmt"
	"net/url"
)

func GetPostgresUrl(user string, password string, host string, port string, database string) string {

	hostPort := fmt.Sprintf("%s:%s", host, port)

	ui := url.UserPassword(user, password)
	// q := fmt.Sprintf("connect_timeout=%d", 30)
	url := url.URL{Scheme: "postgres", Host: hostPort, User: ui, Path: database}

	return url.String()
}

func GetGormPostgresUrl(user string, password string, host string, port string, database string) string {

	formatdsn := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC"
	dsn := fmt.Sprintf(formatdsn, host, user, password, database, port)

	return dsn
}

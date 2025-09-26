# InterSystems IRIS Testcontainers for Go

## Goal

The goal of this project is to make it easy to run tests with InterSystems IRIS Database via Testcontainers.

## Usage

Read to original Testcontainers documentation: https://golang.testcontainers.org/quickstart/


```golang
package tests_test

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	// Import IRIS SQL driver
	_ "github.com/caretdev/go-irisnative"
	iriscontainer "github.com/caretdev/testcontainers-iris-go"
)

var connectionString string = "iris://_SYSTEM:SYS@localhost:1972/USER"

var container *iriscontainer.IRISContainer = nil

func TestMain(m *testing.M) {
	var (
		useContainer   bool
		containerImage string
	)
	flag.BoolVar(&useContainer, "container", true, "Use container image.")
	flag.StringVar(&containerImage, "container-image", "", "Container image.")
	flag.Parse()
	var err error
	ctx := context.Background()
	if useContainer || containerImage != "" {
		options := []testcontainers.ContainerCustomizer{
			iriscontainer.WithNamespace("TEST"),
			iriscontainer.WithUsername("testuser"),
			iriscontainer.WithPassword("testpassword"),
		}
		if containerImage != "" {
			container, err = iriscontainer.Run(ctx, containerImage, options...)
		} else {
			// or use default docker image
			container, err = iriscontainer.RunContainer(ctx, options...)
		}
		if err != nil {
			log.Println("Failed to start container:", err)
			os.Exit(1)
		}
		defer container.Terminate(ctx)
		connectionString = container.MustConnectionString(ctx)
		log.Println("Container started successfully", connectionString)
	}

	var exitCode int = 0

	exitCode = m.Run()

	if container != nil {
		container.Terminate(ctx)
	}
	os.Exit(exitCode)
}

func openDbWrapper[T require.TestingT](t T, dsn string) *sql.DB {
	db, err := sql.Open(`intersystems`, dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

func closeDbWrapper[T require.TestingT](t T, db *sql.DB) {
	if db == nil {
		return
	}
	require.NoError(t, db.Close())
}

func TestOpen(t *testing.T) {
	db := openDbWrapper(t, connectionString)
	defer closeDbWrapper(t, db)

	var (
		namespace string
		username  string
	)
	res := db.QueryRow(`SELECT $namespace, $username`)
	require.NoError(t, res.Scan(&namespace, &username))
	require.Equal(t, "TEST", namespace)
	require.Equal(t, "testuser", username)
}
```

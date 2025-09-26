package iris

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IRISContainer struct {
	testcontainers.Container
	username string
	password string
	namespace string
}

func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["IRIS_USERNAME"] = username

		return nil
	}
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["IRIS_PASSWORD"] = password

		return nil
	}
}

func WithNamespace(namespace string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["IRIS_NAMESPACE"] = namespace

		return nil
	}
}

func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*IRISContainer, error) {
	return Run(ctx, "containers.intersystems.com/intersystems/iris-community:latest-em", opts...)
}

// Run creates an instance of the IRIS container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*IRISContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"1972/tcp", "52773/tcp"},
		Env: map[string]string{
		},
		WaitingFor: wait.ForLog("Enabling logons"),
		AutoRemove: true,
		Cmd: 			[]string{"--after", `iris session iris -U%SYS '##class(Security.Users).UnExpireUserPasswords("*")'`},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	createUser := true
	username, ok := req.Env["IRIS_USERNAME"]
	if !ok {
		username = "_SYSTEM"
		createUser = false
	}
	password, ok := req.Env["IRIS_PASSWORD"]
	if !ok {
		password = "SYS"
	}
	namespace, ok := req.Env["IRIS_NAMESPACE"]
	if !ok {
		namespace = "USER"
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *IRISContainer
	if container != nil {
		c = &IRISContainer{
			Container: container,
			username:  username,
			password:  password,
			namespace:  namespace,
		}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	if err = c.initialize(ctx, createUser); err != nil {
		return c, fmt.Errorf("initialize: %w", err)
	}

	return c, nil
}

func (c *IRISContainer) ConnectionString(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, "1972/tcp", "")
	if err != nil {
		return "", err
	}

	connectionString := fmt.Sprintf("iris://%s:%s@%s/%s", c.username, c.password, endpoint, c.namespace)
	return connectionString, nil
}


func (c *IRISContainer) MustConnectionString(ctx context.Context) string {
	addr, err := c.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	return addr
}


func (c *IRISContainer) initialize(ctx context.Context, createUser bool) error {
	if createUser {
		exitCode, resp, err := c.Exec(ctx, []string{
				"iris",
				"session",
				"iris",
				"-U",
				"%SYS",
				fmt.Sprintf(`##class(Security.Users).Create("%s","%%ALL","%s")`, c.username, c.password),
		})
		if err != nil {
			return fmt.Errorf("exec create user: %w", err)
		}
		if exitCode != 0 {
			return fmt.Errorf("create user failed: %s", resp)
		}
	}

	if c.namespace != "USER" {
		exitCode, resp, err := c.Exec(ctx, []string{
        "iris",
        "session",
        "iris",
        "-U",
        "%SYS",
        fmt.Sprintf(`##class(%%SQL.Statement).%%ExecDirect(,"CREATE DATABASE %s")`, c.namespace),
		})
		if err != nil {
			return fmt.Errorf("exec create database: %w", err)
		}
		if exitCode != 0 {
			return fmt.Errorf("create database failed: %s", resp)
		}
	}
	return nil
}
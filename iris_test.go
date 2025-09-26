package iris_test

import (
	"context"
	"fmt"
	"testing"

	iris "github.com/caretdev/testcontainers-iris-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func TestIRIS(t *testing.T) {
	ctx := context.Background()

	ctr, err := iris.RunContainer(ctx)
	fmt.Println("IRIS is running at:", ctr.MustConnectionString(ctx))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestIRISWithCredentials(t *testing.T) {
	ctx := context.Background()

	ctr, err := iris.RunContainer(ctx, iris.WithUsername("testuser"), iris.WithPassword("testpass"), iris.WithNamespace("TEST"))
	fmt.Println("IRIS is running at:", ctr.MustConnectionString(ctx))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

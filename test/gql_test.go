package test

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	gqlclient "github.com/catalystsquad/protoc-gen-go-ent/client"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

var client = gqlclient.NewClient(http.DefaultClient, "http://localhost:8085/graphql", nil)

func TestCreateDemo(t *testing.T) {
	fake := gqlclient.CreateDemo{CreateDemo: gqlclient.CreateDemo_CreateDemo{}}
	fake.CreateDemo.Name = gofakeit.Name()
	fake.CreateDemo.NumAttendees = int64(gofakeit.Number(0, 1000))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := client.CreateDemo(ctx, fake.CreateDemo.Name, fake.CreateDemo.NumAttendees)
	require.NoError(t, err)
	require.Equal(t, fake.CreateDemo.Name, response.CreateDemo.Name)
	require.Equal(t, fake.CreateDemo.NumAttendees, response.CreateDemo.NumAttendees)
}

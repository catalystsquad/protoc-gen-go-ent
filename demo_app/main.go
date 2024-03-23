package main

import (
	"app/ent"
	"context"
	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

func main() {
	// Create ent.Client and run the schema migration.
	client, err := ent.Open(dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatal("opening ent client", err)
	}
	if err := client.Schema.Create(
		context.Background(),
	); err != nil {
		log.Fatal("opening ent client", err)
	}

	// Configure the server and start listening on :8085.
	srv := handler.NewDefaultServer(NewSchema(client))
	srv.Use(entgql.Transactioner{TxOpener: client})
	http.Handle("/",
		playground.Handler("Todo", "/graphql"),
	)
	http.Handle("/graphql", srv)
	log.Println("listening on :8085")
	if err := http.ListenAndServe(":8085", nil); err != nil {
		log.Fatal("http server terminated", err)
	}
}

package inits

import (
	"fmt"
	"forum/auth"
	"forum/db/DAO"
	"forum/graph"
	"forum/handlers"
	"forum/loaders"
	"log"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/vektah/gqlparser/v2/ast"
)

// Здесь происходит вся настройка роутера.
func InitRouter(dao DAO.Dao, errorsLog *log.Logger, dbErrorsLog *log.Logger) (router *chi.Mux) {
	router = chi.NewRouter()

	// Подключение Middleware
	router.Use(auth.AuthMiddleware(dao))
	router.Use(loaders.CommentsLoaderMiddleware(dao, dbErrorsLog))

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Dao: dao, ErrorsLog: errorsLog, DbErrorsLog: dbErrorsLog}}))

	srv.AddTransport(transport.Websocket{})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	h := handlers.CreateHandlers(dao, errorsLog, dbErrorsLog)

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)
	router.HandleFunc("/registration", h.Registration)
	router.HandleFunc("/registration/form", h.RegistrationForm)
	router.HandleFunc("/login", h.Login)
	router.HandleFunc("/login/form", h.LoginForm)

	fmt.Printf("%-25s %-25s\n", "Маршрут", "Описание")
	fmt.Println("--------------------------------------")
	fmt.Printf("%-25s %-25s\n", "/", "GraphQL playground")
	fmt.Printf("%-25s %-25s\n", "/query", "GraphQL playground")
	fmt.Printf("%-25s %-25s\n", "/registration", "Регистрация (принимаются и json поля username и password)")
	fmt.Printf("%-25s %-25s\n", "/registration/form", "Форма регистрации")
	fmt.Printf("%-25s %-25s\n", "/login", "Вход в систему (принимаются и json поля username и password)")
	fmt.Printf("%-25s %-25s\n", "/login/form", "Форма входа в систему")
	return
}

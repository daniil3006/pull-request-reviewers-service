package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"pull-request-reviewers-service/internal/api"
	"pull-request-reviewers-service/internal/repository"
	"pull-request-reviewers-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DB *pgxpool.Pool
}

func (s *Server) Init() {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	db := os.Getenv("POSTGRES_DB")

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, db)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("postgres connect error")
	}

	if err = pool.Ping(context.Background()); err != nil {
		log.Fatal("postgres ping error")
	}
	log.Println("Success connest to PostgreSQL")

	s.DB = pool
}

func (s *Server) Start() {
	teamRepo := repository.NewTeamRepository(s.DB)
	teamService := service.NewTeamService(teamRepo)
	teamHandler := api.NewTeamHandler(teamService)

	prRepo := repository.NewPullRequestRepository(s.DB)
	prService := service.NewPullRequestService(prRepo, teamRepo)
	prHandler := api.NewPullRequestHandler(prService)

	r := chi.NewRouter()
	r.Post("/team/add", teamHandler.CreateTeam)
	r.Get("/team/get", teamHandler.GetTeam)
	r.Post("/users/setIsActive", teamHandler.SetIsActiveUser)
	r.Post("/pullRequest/create", prHandler.CreatePullRequest)
	r.Post("/pullRequest/merge", prHandler.MergePullRequest)
	r.Post("/pullRequest/reassign", prHandler.ReassignReviewer)
	r.Get("/users/getReview", teamHandler.GetPRsByReviewer)
	r.Get("/stats/reviewers", prHandler.GetAssignStat)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("error connect")
	}
}

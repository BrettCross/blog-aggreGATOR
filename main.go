package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/brettcross/blog-aggreGATOR/internal/config"
	"github.com/brettcross/blog-aggreGATOR/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	
	db, err := sql.Open("postgres", cfg.DBURL)
	dbQueries := database.New(db)
	
	programState := &state{
		db: dbQueries,
		cfg: &cfg,
	}

	cmds := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = cmds.run(programState, command{Name: cmdName, Args: cmdArgs})
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("Read config: %+v\n", cfg)

	// err = cfg.SetUser("Brett")
	// if err != nil {
	// 	log.Fatalf("couldn't set current user :%v", err)
	// }

	// cfg, err = config.Read()
	// if err != nil {
	// 	log.Fatalf("error reading config: %v", err)
	// }
	// fmt.Printf("Read config again: %+v\n", cfg)

}
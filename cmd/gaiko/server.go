package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

type ProveData struct {
	ProveMode string `json:"prove_mode"` // block, batch, aggregation
	Input     []byte `json:"input,omitempty"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Proof   string `json:"proof"`
}

type ProveMode int

const (
	OntakeBlock   ProveMode = iota // 0
	PacayaBatch                    // 1
	Aggregation                    // 2
	Bootstrap                      // 3
	Check                          // 4
	TestHeartBeat                  // 5
	HeklaBlock                     // 6 deprecated
)

func proveHandler(ctx context.Context, args *flags.Arguments, sgxProver *prover.SGXProver, w http.ResponseWriter, r *http.Request, proveMode ProveMode) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST Only", http.StatusMethodNotAllowed)
		return
	}

	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, "Read request failed", http.StatusBadRequest)
	// 	return
	// }
	defer r.Body.Close()
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		fmt.Printf("Prove recievied content type: %s\n", contentType)
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// call different command according to data.Type
	var err error
	switch proveMode {
	case TestHeartBeat:
		fmt.Fprintf(args.ProofWriter, "Hello, %s!", "world")
	case OntakeBlock:
		err = oneshot(ctx, sgxProver, args)
	case HeklaBlock:
		http.Error(w, "Hekla block prove is deprecated", http.StatusBadRequest)
		err = fmt.Errorf("hekla block prove is deprecated")
	case PacayaBatch:
		err = batchOneshot(ctx, sgxProver, args)
	case Aggregation:
		err = aggregate(ctx, sgxProver, args)
	case Bootstrap:
		err = bootstrap(ctx, sgxProver, args)
	case Check:
		err = check(ctx, sgxProver, args)
	default:
		http.Error(w, "Unknown prove mode", http.StatusBadRequest)
		err = fmt.Errorf("unknown prove mode: %d", proveMode)
	}
	var status string = "success"
	var message string
	if err != nil {
		message = err.Error()
		status = "failed"
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Println("Prove finished, get proof %s", args.ProofWriter.(*bytes.Buffer).String())
	response := Response{
		Status:  status,
		Message: message,
		Proof:   args.ProofWriter.(*bytes.Buffer).String(),
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Response serialize failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func runServer(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8080"
	}

	// TODO: using closet for easy test maybe not a good idea
	http.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		args := flags.NewArguments(c)
		// override the proof writer to get the proof & return as response
		args.ProofWriter = new(bytes.Buffer)
		args.WitnessReader = r.Body
		sgxProver := prover.NewSGXProver(args)
		proveHandler(c.Context, args, sgxProver, w, r, OntakeBlock)
	})
	http.HandleFunc("/batch", func(w http.ResponseWriter, r *http.Request) {
		args := flags.NewArguments(c)
		// override the proof writer to get the proof & return as response
		args.ProofWriter = new(bytes.Buffer)
		args.WitnessReader = r.Body
		sgxProver := prover.NewSGXProver(args)
		proveHandler(c.Context, args, sgxProver, w, r, PacayaBatch)
	})
	http.HandleFunc("/aggregation", func(w http.ResponseWriter, r *http.Request) {
		args := flags.NewArguments(c)
		// override the proof writer to get the proof & return as response
		args.ProofWriter = new(bytes.Buffer)
		args.WitnessReader = r.Body
		sgxProver := prover.NewSGXProver(args)
		proveHandler(c.Context, args, sgxProver, w, r, PacayaBatch)
	})
	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		args := flags.NewArguments(c)
		// override the proof writer to get the proof & return as response
		args.ProofWriter = new(bytes.Buffer)
		args.WitnessReader = r.Body
		sgxProver := prover.NewSGXProver(args)
		proveHandler(c.Context, args, sgxProver, w, r, PacayaBatch)
	})
	http.HandleFunc("/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		args := flags.NewArguments(c)
		// override the proof writer to get the proof & return as response
		args.ProofWriter = new(bytes.Buffer)
		sgxProver := prover.NewSGXProver(args)
		proveHandler(c.Context, args, sgxProver, w, r, Bootstrap)
	})

	server := &http.Server{
		Addr: ":" + port,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nServer is closing...")
		server.Close()
	}()

	fmt.Printf("Start server @ http://localhost:%s\n", port)
	return server.ListenAndServe()
}

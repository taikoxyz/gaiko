package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/taikoxyz/gaiko/internal/flags"
	"github.com/taikoxyz/gaiko/internal/prover"
	"github.com/urfave/cli/v2"
)

var bytesBufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

type ProveData struct {
	ProveMode string `json:"prove_mode"` // block, batch, aggregation
	Input     []byte `json:"input,omitempty"`
}

type Response struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Proof   json.RawMessage `json:"proof"`
}

type ProveMode int

const (
	Unknown       ProveMode = iota // 0
	OntakeBlock                    // 1
	PacayaBatch                    // 2
	Aggregation                    // 3
	Bootstrap                      // 4
	StatusCheck                    // 5
	TestHeartBeat                  // 6
	HeklaBlock                     // 7 deprecated
)

func proveHandler(ctx context.Context, args *flags.Arguments, sgxProver *prover.SGXProver, w http.ResponseWriter, r *http.Request, proveMode ProveMode) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST Only", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		fmt.Printf("Prove recievied content type: %s\n", contentType)
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// call different command according to data.Type
	var err error
	var proofResponse []byte
	switch proveMode {
	case TestHeartBeat:
		fmt.Fprintf(args.ProofWriter, "Hello, %s!", "world")
	case OntakeBlock:
		err = oneshot(ctx, sgxProver, args)
		proofResponse = args.ProofWriter.(*bytes.Buffer).Bytes()
	case HeklaBlock:
		http.Error(w, "Hekla block prove is deprecated", http.StatusBadRequest)
		return
	case PacayaBatch:
		err = batchOneshot(ctx, sgxProver, args)
		proofResponse = args.ProofWriter.(*bytes.Buffer).Bytes()
	case Aggregation:
		err = aggregate(ctx, sgxProver, args)
		proofResponse = args.ProofWriter.(*bytes.Buffer).Bytes()
	case Bootstrap:
		err = bootstrap(ctx, sgxProver, args)
	case StatusCheck:
		err = check(ctx, sgxProver, args)
	default:
		http.Error(w, "Unknown prove mode", http.StatusBadRequest)
		return
	}

	var response Response
	if err != nil {
		fmt.Printf("Prove finished, get error: %s, response: %s\n", err.Error(), string(proofResponse))
		response = Response{
			Status:  "error",
			Message: err.Error(),
			Proof:   []byte("{}"),
		}
	} else {
		fmt.Printf("Prove finished, get proof %s\n", proofResponse)
		response = Response{
			Status:  "success",
			Message: "",
			Proof:   proofResponse,
		}
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
	http.HandleFunc("POST /prove/{action}", func(w http.ResponseWriter, r *http.Request) {
		args := flags.NewArguments(c)
		// override the proof writer to get the proof & return as response
		args.ProofWriter = bytesBufferPool.Get().(*bytes.Buffer)
		defer bytesBufferPool.Put(args.ProofWriter)
		args.WitnessReader = r.Body
		sgxProver := prover.NewSGXProver(args)
		proveMode := Unknown
		if r.URL.Query().Get("debug") == "true" {
			args.SGXType = "debug"
		}
		switch r.PathValue("action") {
		case "block":
			proveMode = OntakeBlock
		case "batch":
			proveMode = PacayaBatch
		case "aggregate":
			proveMode = Aggregation
		case "bootstrap":
			proveMode = Bootstrap
		case "check":
			proveMode = StatusCheck
		case "heartbeat":
			proveMode = TestHeartBeat
		case "hekla":
			proveMode = HeklaBlock
		}
		proveHandler(r.Context(), args, sgxProver, w, r, proveMode)
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

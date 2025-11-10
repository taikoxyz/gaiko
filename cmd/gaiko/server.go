package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
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

type ProveMode string

const (
	Unknown         ProveMode = "unknown"
	OntakeBlock     ProveMode = "block"
	PacayaBatch     ProveMode = "batch"
	Aggregation     ProveMode = "aggregate"
	ShastaAggregate ProveMode = "shasta-aggregate"
	Bootstrap       ProveMode = "bootstrap"
	StatusCheck     ProveMode = "check"
	TestHeartBeat   ProveMode = "heartbeat"
	HeklaBlock      ProveMode = "hekla" // deprecated
)

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func proveHandler(ctx context.Context, args *flags.Arguments, sgxProver *prover.SGXProver, w http.ResponseWriter, r *http.Request, proveMode ProveMode) {
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
		return
	case PacayaBatch:
		err = batchOneshot(ctx, sgxProver, args)
	case Aggregation:
		err = aggregate(ctx, sgxProver, args)
	case ShastaAggregate:
		err = shastaAggregate(ctx, sgxProver, args)
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
		log.Debug("Prove finished, get error: %s, response: ", "error", err, "proof", args.ProofWriter.(*bytes.Buffer).String())
		response = Response{
			Status:  "error",
			Message: err.Error(),
			Proof:   []byte("{}"),
		}
	} else {
		log.Debug("Prove finished, get proof: ", "proof", args.ProofWriter.(*bytes.Buffer).String())
		response = Response{
			Status:  "success",
			Message: "",
			Proof:   args.ProofWriter.(*bytes.Buffer).Bytes(),
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
	args := flags.NewArguments(c)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /prove/{action}", func(w http.ResponseWriter, r *http.Request) {
		args := args.Copy()
		defer r.Body.Close()
		// override the proof writer to get the proof & return as response
		buf := bytesBufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		args.ProofWriter = buf
		defer bytesBufferPool.Put(args.ProofWriter)
		args.WitnessReader = r.Body
		sgxProver := prover.NewSGXProver(args)
		proveMode := Unknown
		if r.URL.Query().Get("debug") == "true" {
			args.SGXType = "debug"
			if r.URL.Query().Get("sgx_instance") != "" {
				args.SGXInstance = common.HexToAddress(r.URL.Query().Get("sgx_instance"))
			}
		}
		if r.PathValue("action") != "" {
			proveMode = ProveMode(r.PathValue("action"))
		}
		proveHandler(r.Context(), args, sgxProver, w, r, proveMode)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: recoverMiddleware(mux),
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Info("\nServer is closing...")
		server.Close()
	}()

	log.Info("Start server @ http://[::]: ", "port", port)
	return server.ListenAndServe()
}

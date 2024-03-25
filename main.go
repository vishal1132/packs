package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/vishal1132/packs/config"
)

const (
	defaultPort = 3000
)

type app struct {
	packs []int
	lock  sync.RWMutex
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.LoadConfig()
	a := app{
		packs: []int{250, 500, 1000, 2000, 5000}, // starting with these.
	}
	http.HandleFunc("/pack-sizes", a.removeSizeInputHandler)
	http.HandleFunc("/calculate-packs", a.calculatePacksHandler)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":3000", nil)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", defaultPort),
		Handler:      nil,
		WriteTimeout: time.Duration(cfg.AppConfig.ServerWriteTimeout) * time.Millisecond,
		ReadTimeout:  time.Duration(cfg.AppConfig.ServerWriteTimeout) * time.Millisecond,
	}
	go srv.ListenAndServe()
	waitForShutdownSignal(ctx, cfg.AppConfig.GracefulShutDownTimeout, srv, cancel)
}

type packs struct {
	Size int `json:"size"`
}

func (a *app) removeSizeInputHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	a.lock.Lock()
	defer func() {
		a.lock.Unlock()
	}()
	packs := new(packs)
	if err := json.NewDecoder(r.Body).Decode(packs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	size := packs.Size
	if size <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		for i, v := range a.packs {
			if v == size {
				a.packs = append(a.packs[:i], a.packs[i+1:]...)
			}
		}
	case http.MethodPost:
		if len(a.packs) == 0 || size > a.packs[len(a.packs)-1] {
			a.packs = append(a.packs, size)
			return
		}
		if size < a.packs[0] {
			a.packs = append([]int{size}, a.packs...)
		}
		for i, v := range a.packs {
			if v == packs.Size {
				return
			}
			if i == len(a.packs)-1 {
				return
			}
			if size > v && size < a.packs[i+1] {
				a.packs = append(a.packs, 0)
				copy(a.packs[i+2:], a.packs[i+1:])
				a.packs[i+1] = size
			}
		}
	}
}

type PacksResponse struct {
	Data map[int]int `json:"data"`
}

type PackCalculationRequest struct {
	OrderSize int `json:"orderSize"`
}

func (a *app) calculatePacksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	a.lock.RLock()
	defer a.lock.RUnlock()

	if len(a.packs) == 0 {
		resp := new(PacksResponse)
		json.NewEncoder(w).Encode(resp)
		return
	}

	req := new(PackCalculationRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resp := new(PacksResponse)
	resp.Data = make(map[int]int)
	remainingItems := req.OrderSize
	for i := len(a.packs) - 1; i >= 0; i-- {
		packSize := a.packs[i]
		packs := remainingItems / packSize
		if packs > 0 {
			resp.Data[packSize] = packs
			remainingItems -= packs * packSize
		}
	}

	// Add extra pack if necessary
	if remainingItems > 0 {
		resp.Data[a.packs[0]] = resp.Data[a.packs[0]] + 1
	}
	if len(a.packs) > 1 {
		if _, ok := resp.Data[a.packs[0]]; ok {
			divideFactor := a.packs[1] / a.packs[0]
			if a.packs[1]%a.packs[0] > 0 {
				divideFactor = 0
			}
			if divideFactor != 0 {
				initialPack := resp.Data[a.packs[0]]
				resp.Data[a.packs[0]] = resp.Data[a.packs[0]] - ((resp.Data[a.packs[0]] / divideFactor) * divideFactor)
				resp.Data[a.packs[1]] = resp.Data[a.packs[1]] + initialPack/divideFactor
			}
		}
	}
	for i, v := range resp.Data {
		if v == 0 {
			delete(resp.Data, i)
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func waitForShutdownSignal(ctx context.Context, gracefulShutDownTimeout int64, srv *http.Server, cancel context.CancelFunc) {
	var gracefulStop = make(chan os.Signal, 1)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGQUIT)

	select {
	case <-gracefulStop:
		cancel()
		// if stop signal is received, wait for some time so that background workers get time to exit
		<-time.After(time.Duration(gracefulShutDownTimeout) * time.Millisecond)
	case <-ctx.Done():
		// shutdown if context was cancelled by something else before shutdown signal
	}
	srv.Shutdown(ctx)
}

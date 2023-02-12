package main

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

// listenAndServeCtx wraps srv.ListenAndServe with a context.Context.
func listenAndServeCtx(ctx context.Context, srv *http.Server) error {
	errC := make(chan error)
	go func() {
		errC <- srv.ListenAndServe()
	}()
	select {
	case err := <-errC:
		return err
	case <-ctx.Done():
		// Intentionally using context.Background() because ctx is already cancelled.
		timeout, canc := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer canc()
		_ = srv.Shutdown(timeout)
		return ctx.Err()
	}
}

func multiListen(ctx context.Context, hdl http.Handler, addrs []string) error {
	eg, ctx := errgroup.WithContext(ctx)
	for _, addr := range addrs {
		addr := addr // copy
		eg.Go(func() error {
			srv := &http.Server{
				Handler: hdl,
				Addr:    addr,
			}
			return listenAndServeCtx(ctx, srv)
		})
	}
	return eg.Wait()
}

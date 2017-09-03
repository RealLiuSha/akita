// +build go1.8

package akita

import (
	stdContext "context"
)

// Close immediately stops the server.
// It internally calls `http.Server#Close()`.
func (a *Akita) Close() error {
	if err := a.TLSServer.Close(); err != nil {
		return err
	}
	return a.Server.Close()
}

// Shutdown stops server the gracefully.
// It internally calls `http.Server#Shutdown()`.
func (a *Akita) Shutdown(ctx stdContext.Context) error {
	if err := a.TLSServer.Shutdown(ctx); err != nil {
		return err
	}
	return a.Server.Shutdown(ctx)
}

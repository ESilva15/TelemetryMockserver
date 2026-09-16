// Package mockserver is the core of this program and it will have the API
// to record binary data from the UDP server and then be able to mock it
package mockserver

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sync"
	"time"

	"github.com/ESilva15/TelemetryMockserver/constants"
	bngsdk "github.com/ESilva15/gobngsdk"
)

type ViewData struct {
	Outgauge bngsdk.Outgauge
	SizeRead int64
	FileSize int64
}

// Replayer does the replaying
// Should we make a "player" struct that can record and replay?
type Replayer struct {
	SDK *bngsdk.BeamNGSDK

	// Streams
	dataViewCh chan ViewData

	// Mut
	mut sync.RWMutex

	// View
	viewData ViewData
}

func NewReplayer(address string, port int, fp string) (*Replayer, error) {
	sdk, err := bngsdk.NewBngSDK(bngsdk.Options{
		Logger:           slog.Default().With("SDK", "BeamNG"),
		SourceType:       bngsdk.BinaryFile,
		BinSourcePath:    fp,
		ExportData:       true,
		ExportDataType:   bngsdk.UDPData,
		ExportUDPAddress: address,
		ExportUDPPort:    port,
		Loop:             true,
	})
	if err != nil {
		return nil, err
	}

	replayer := &Replayer{
		SDK:        sdk,
		dataViewCh: make(chan ViewData, 1),
	}

	return replayer, nil
}

// renderToTerminal will render the data for the users viewing pleasure
func (r *Replayer) renderToTerminal(ctx context.Context) {
	var buf bytes.Buffer
	buf.Grow(2048)

	for {
		select {
		case <-ctx.Done():
			return
		case data := <-r.dataViewCh:
			// Reset to the start of the terminal
			// percent := int(float64(data.SizeRead) / float64(data.FileSize) * 100)
			percent := int(math.Round((float64(data.SizeRead) / float64(data.FileSize)) * 100))

			buf.Reset()
			buf.WriteString("\x1b[2J\x1b[H")
			fmt.Fprintf(&buf, "\x1b]0;%s - Replaying %d%%\x07", constants.ProgramName, percent)

			fmt.Fprintf(&buf, "Replayed: %d%%\n", percent)

			stringifyOutgaugeData(&buf, &data.Outgauge)

			buf.WriteTo(os.Stdout)
		}
	}
}

// Replay replays a given file <fp> in a UDP server <addr>:<port>
func (r *Replayer) Replay(ctx context.Context, loop bool) error {
	go r.renderToTerminal(ctx)

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			og, err := r.SDK.Update()
			if err != nil {
				// TODO: log here
				slog.Error("an error occurred when updating", "err", err)
				return err
			}

			r.mut.Lock()
			r.viewData.SizeRead = r.SDK.GetTotalRead()
			r.viewData.FileSize = r.SDK.GetSourceSize()
			r.viewData.Outgauge = *og
			r.mut.Unlock()

			// Send the data to the view
			select {
			case r.dataViewCh <- r.viewData:
				// Sent the data
			default:
				// Dropped the frame!
			}
		}
	}
}

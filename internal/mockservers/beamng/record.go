package mockserver

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/ESilva15/TelemetryMockserver/constants"
	bngsdk "github.com/ESilva15/gobngsdk"
)

type recorderViewData struct {
	TotalBytes int64
	Og         bngsdk.Outgauge
}

type Recorder struct {
	SDK *bngsdk.BeamNGSDK
	// Views
	mut         sync.RWMutex
	viewDataMut sync.RWMutex
	viewData    recorderViewData
	viewCh      chan *recorderViewData
}

func NewRecorder(fp string, address string, port int) (*Recorder, error) {
	var recorder Recorder
	var err error

	recorder.SDK, err = bngsdk.NewBngSDK(bngsdk.Options{
		Logger:           slog.Default().With("SDK", "BeamNG"),
		SourceType:       bngsdk.UDPData,
		ImportUDPAddress: address,
		ImportUDPPort:    port,
		ExportData:       true,
		ExportDataType:   bngsdk.BinaryFile,
		ExportDataPath:   fp,
	})
	if err != nil {
		return nil, err
	}

	recorder.viewData = recorderViewData{}
	recorder.viewCh = make(chan *recorderViewData, 1)

	return &recorder, nil
}

func (r *Recorder) Close() {
	r.SDK.Close()
}

func (r *Recorder) view(ctx context.Context) {
	var buf bytes.Buffer

	buf.Grow(2048)

	for {
		select {
		case <-ctx.Done():
			return
		case viewData := <-r.viewCh:
			buf.Reset()
			buf.WriteString("\x1b[2J\x1b[H")
			fmt.Fprintf(&buf, "\x1b]0;%s - Recording ", constants.ProgramName)
			stringifyRecordingProgress(&buf, viewData.TotalBytes)
			fmt.Fprintf(&buf, "\x07")

			stringifyRecordingProgress(&buf, viewData.TotalBytes)
			fmt.Fprintf(&buf, "\n\n")

			r.viewDataMut.RLock()
			stringifyOutgaugeData(&buf, &viewData.Og)
			r.viewDataMut.RUnlock()

			_, _ = buf.WriteTo(os.Stdout)
		}
	}
}

// Record records data from the UDP connection created by address and port
func (r *Recorder) Record(ctx context.Context) error {
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	go r.view(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			og, err := r.SDK.Update()
			if err != nil {
				return err
			}

			r.viewData.TotalBytes = r.SDK.GetTotalWritten()
			r.viewData.Og = *og

			// Send the data to the view
			select {
			case r.viewCh <- &r.viewData:
				// Sent the data
			default:
				// Dropped the frame!
			}
		}
	}
}

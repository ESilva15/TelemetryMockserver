package iracing

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ESilva15/goirsdk"
)

type Replayer struct {
	SDK *goirsdk.IBT
	// DataSourcePath string
	// Socket         *UDPTransport

	// Output memory mapped file
	outputMmap string

	// Streams
	// dataViewCh chan ViewData
	// socketCh   chan []byte

	// Mut
	mut sync.RWMutex

	// View
	// viewData ViewData
}

func NewReplayer(input string, output string) (*Replayer, error) {
	ibt, err := goirsdk.Init(goirsdk.Options{
		SourceType:    goirsdk.IBTFile,
		SourcePath:    "../testTelemetry/gt3_mustang_bathurst.ibt",
		IBTExportType: goirsdk.SharedMemoryFile,
		IBTExport:     true,
	})
	if err != nil {
		return nil, err
	}

	replayer := &Replayer{
		SDK: ibt,
	}

	return replayer, nil
}

func (r *Replayer) Replay(ctx context.Context, loop bool) error {
	mainLoopTicker := time.NewTicker(time.Second / 240)
	defer mainLoopTicker.Stop()

	for {
		// Update the data that the SDK is holding with the next tick
		_, err := r.SDK.Update(100 * time.Millisecond)
		if err != nil {
			log.Printf("could not update data: %v", err)
			continue
		}

		// Vehicle Movement data gathered from the names we can find on the
		// telemetry_docs.pdf file
		// - I wish to make this less verbose if possible
		if _, ok := r.SDK.Vars.Vars["Gear"]; !ok {
			log.Fatal("Field `Gear` doesn't exist")
		}

		if _, ok := r.SDK.Vars.Vars["RPM"]; !ok {
			log.Fatal("Field `RPM` doesn't exist")
		}

		if _, ok := r.SDK.Vars.Vars["Speed"]; !ok {
			log.Fatal("Field `Speed` doesn't exist")
		}

		gear := int32(r.SDK.Vars.Vars["Gear"].Value.(int))
		rpm := int32(r.SDK.Vars.Vars["RPM"].Value.(float32))
		speed := int32(r.SDK.Vars.Vars["Speed"].Value.(float32))

		fmt.Printf("\033[?25l\033[2J\033[H")
		fmt.Printf("Gear: %d, RPM: %d, Speed: %d", gear, rpm, speed)

		<-mainLoopTicker.C
	}
}

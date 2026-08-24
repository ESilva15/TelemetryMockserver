package cmd

import (
	"context"
	"fmt"

	"github.com/ESilva15/TelemetryMockserver/internal/mockservers/iracing"
	"github.com/spf13/cobra"
)

// beamNGCmd is the parent command for the BeamNG agent actions
var iracingCmd = &cobra.Command{
	Use:   "iracing",
	Short: "iRacing telemetry utility",
	Args:  nil,
}

var iracingReplayCMD = &cobra.Command{
	Use:   "replay",
	Short: "replay -i <input-file> -o <output-memory-map-file>",
	Long:  "replays the <input-file> in the <output-memory-map-file>",
	Args:  nil,
	Run:   iracingReplayAction,
}

func iracingReplayAction(cmd *cobra.Command, args []string) {
	// We need to get the path of the ibt.file
	// outputFile, _ := cmd.Flags().GetString("output")
	// inputFile, _ := cmd.Flags().GetString("input")

	replayer, err := iracing.NewReplayer("", "")
	if err != nil {
		fmt.Printf("Something went wrong setting up the player: %+v", err)
		return
	}

	// NOTE: is this doing anything at all??
	ctx := context.Background()
	if err := replayer.Replay(ctx, false); err != nil {
		fmt.Printf("Something went wrong while playing the file: %v", err)
	}
}

func init() {
	rootCmd.AddCommand(iracingCmd)

	iracingCmd.AddCommand(iracingReplayCMD)

	// iracingReplayCMD flags
	iracingReplayCMD.Flags().StringP("input", "i", "input.ibt", "input file")
	// iracingReplayCMD.MarkFlagRequired("input")
	iracingReplayCMD.Flags().StringP("output", "o", "memmap.file", "output memory mapped file")
	// iracingReplayCMD.MarkFlagRequired("output")
}

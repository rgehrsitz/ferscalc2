package cli

import (
	"fmt"
	"log"

	"github.com/rpgo/retirement-calculator/internal/web"
	"github.com/spf13/cobra"
)

var port int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web interface",
	Long: `Starts a local web server that provides a graphical interface
for the FERS retirement calculator.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting web server on http://localhost:%d\n", port)
		if err := web.StartServer(port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
}

package cmd

import (
	"database/sql"
	"github.com/edwardsb/secureworks/geoip"
	"github.com/edwardsb/secureworks/internal/httpd"
	"github.com/edwardsb/secureworks/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the Superman Detector Web Service",
	Run: func(cmd *cobra.Command, args []string) {

		type Module interface {
			Open() error
			Close() error
		}

		// start creating dependencies
		geoip := geoip.NewService(viper.GetString("GEOLITE_PATH"))
		db, err := sql.Open("sqlite3", viper.GetString("DB_PATH"))
		if err != nil {
			log.Panic(err)
		}
		// start injecting dependencies
		store := store.NewSqliteDb(db)
		httpServer := httpd.NewHTTPServer(store, geoip)


		modules := []Module{geoip, httpServer, store}
		for _, m := range modules {
			err := m.Open()
			if err != nil {
				log.Fatal(err)
			}
		}


		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

		shutDownComplete := make(chan struct{})
		go func() {
			<- sigChan
			log.Println("closing modules")
			for _, m := range modules {
				err := m.Close()
				if err != nil {
					log.Fatal(err)
				}
			}
			close(shutDownComplete)
		}()

		<- shutDownComplete
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

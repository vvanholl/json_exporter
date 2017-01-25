package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"

	"github.com/vvanholl/json_exporter/collector"
)

func main() {
	var (
		config        *collector.Config
		exporter      *collector.Exporter
		configFile    = flag.String("path.config", "config.yml", "Configuration file.")
		listenAddress = flag.String("web.listen-address", ":8888", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		showVersion   = flag.Bool("version", false, "Print version information.")
		err           error
	)
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("json_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting json_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	config, err = collector.NewFileConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	exporter, err = collector.NewExporter(config)
	if err != nil {
		log.Fatal("Could not start exporter", err)
	}

	exporter.StartRoutines()

	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("json_exporter"))

	log.Infoln("Listening on", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>JSON Exporter</title></head>
             <body>
             <h1>JSON Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

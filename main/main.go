package main

import (
	"encoding/json"
	"github.com/codegangsta/cli"
	"github.com/xytis/graylog-golang"
	"github.com/xytis/j2g/journal"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	var flagLogLevel = cli.StringFlag{
		Name:  "log-level",
		Value: "info",
		Usage: "logging level (debug, info, warning, error)",
	}

	var flagGelfHost = cli.StringFlag{
		Name:  "gelf-host",
		Value: "127.0.0.1",
		Usage: "gelf endpoint hostname",
	}

	var flagGelfPort = cli.IntFlag{
		Name:  "gelf-port",
		Value: 12201,
		Usage: "gelf endpoint port",
	}

	var flagGelfConnection = cli.StringFlag{
		Name:  "gelf-connection",
		Value: "wan",
		Usage: "gelf connection type (wan, lan)",
	}

	var flagGelfMaxChunkSizeWan = cli.IntFlag{
		Name:  "gelf-max-chunk-size-wan",
		Value: 1420,
		Usage: "gelf max chunk size for wan connection",
	}

	var flagGelfMaxChunkSizeLan = cli.IntFlag{
		Name:  "gelf-max-chunk-size-lan",
		Value: 8154,
		Usage: "gelf max chunk size for lan connection",
	}

	app := cli.NewApp()
	app.Name = "j2g"
	app.Usage = "journald forwarder to gelf endpoint"
	app.Version = Version
	app.Flags = []cli.Flag{
		flagLogLevel,
		flagGelfHost,
		flagGelfPort,
		flagGelfConnection,
		flagGelfMaxChunkSizeWan,
		flagGelfMaxChunkSizeLan,
	}

	app.Action = Run
	app.Run(os.Args)
}

func Run(ctx *cli.Context) {
	SetLogLevel(ctx.String("log-level"))

	j, err := journal.NewJournal()
	if err != nil {
		panic(err)
	}

	g := gelf.New(gelf.Config{
		GraylogPort:     ctx.Int("gelf-port"),
		GraylogHostname: ctx.String("gelf-host"),
		Connection:      ctx.String("gelf-connection"),
		MaxChunkSizeWan: ctx.Int("gelf-max-chunk-size-wan"),
		MaxChunkSizeLan: ctx.Int("gelf-max-chunk-size-lan"),
	})

	start := time.Now()
	if err := j.SeekRealtimeUsec(uint64(start.UnixNano() / 1000)); err != nil {
		panic(err)
	}

	Log.Infoln("Starting journal")
	until := make(chan struct{})
	event := make(chan int, 1)
	done := make(chan struct{}, 1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		until <- struct{}{}
	}()
listen:
	for {
		c, err := j.Next()
		if err != nil {
			Log.Errorf("Journal traversing error %v\n:", err)
			continue
		}
		if c == 1 {
			e, err := j.GetEntry()
			if err != nil {
				Log.Errorf("Skiping unreadable entry: %v\n", err)
				continue
			}
			Log.Debugf("Received entry: %v\n", e)
			raw, err := json.Marshal(e)
			if err != nil {
				Log.Errorf("Skipping unserializable entry: %v\n", e)
				continue
			}
			g.RawLog(raw)
		}
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					event <- j.Wait(time.Duration(1) * time.Second)
				}
			}
		}()
		select {
		case <-until:
			done <- struct{}{}
			break listen
		case e := <-event:
			done <- struct{}{}
			switch e {
			case journal.SD_JOURNAL_NOP, journal.SD_JOURNAL_APPEND, journal.SD_JOURNAL_INVALIDATE:
				// TODO: need to account for any of these?
				// https://www.freedesktop.org/software/systemd/man/sd_journal_wait.html#
			default:
				Log.Warnf("Received unknown event: %d\n", e)
			}
			continue
		}
	}
	Log.Infoln("Closing journal")
	j.Close()
}

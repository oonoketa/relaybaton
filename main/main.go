package main

import (
	"github.com/iyouport-org/relaybaton"
	"github.com/iyouport-org/relaybaton/dns"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	err := os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1,netdns=go")
	if err != nil {
		log.Fatal(err)
		return
	}
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		log.Error(err)
		return
	}
	var conf relaybaton.Config
	if err := v.Unmarshal(&conf); err != nil {
		log.Error(err)
		return
	}
	file, err := os.OpenFile(conf.LogFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err)
	}
	log.SetOutput(file)
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(relaybaton.XMLFormatter{})
	log.SetReportCaller(true)

	if conf.Client.DoH == "dot" {
		net.DefaultResolver = dns.NewDoTResolverFactory(net.Dialer{}, "cloudflare-dns.com", "1.0.0.1:853", false).GetResolver()
	}

	switch os.Args[1] {
	case "client":
		for {
			client, err := relaybaton.NewClient(conf)
			if err != nil {
				log.Error(err)
				time.Sleep(5 * time.Second)
				continue
			}
			client.Run()
			time.Sleep(5 * time.Second)
		}
	case "server":
		handler := relaybaton.Handler{
			Conf: conf,
		}
		log.Error(http.ListenAndServe(":"+strconv.Itoa(conf.Server.Port), handler))
	}
}

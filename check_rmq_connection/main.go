package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	RabbitMQConnectionMetric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rabbitmq_connection_exit_code",
		Help: "RabbitMQ connection exit code. 0 - successful or 1 - failed",
	})

	rmq_user     = os.Getenv("RMQ_USER")
	rmq_password = os.Getenv("RMQ_PASSWORD")
	rmq_host     = os.Getenv("RMQ_HOST")
	rmq_vhost    = os.Getenv("RMQ_VHOST")
	rmq_port     = os.Getenv("RMQ_PORT")
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func connect() float64 {
	connstring := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", rmq_user, rmq_password, rmq_host, rmq_port, rmq_vhost)

	if conn, err := amqp.Dial(connstring); err != nil {
		return 1
	} else {
		defer conn.Close()
		if ch, err := conn.Channel(); err != nil {
			return 1
		} else {
			defer ch.Close()
		}
	}

	return 0
}

func main() {
	okCount := 0
	errCount := 0
	summCount := 0

	go func() {
		for {
			exitCode := connect()
			if exitCode == 0 {
				okCount += 1
			}
			if exitCode == 1 {
				errCount += 1
			}
			summCount += 1
			if summCount == 200 {
				fmt.Println("Successful connections within one hour: ", okCount)
				fmt.Println("Connection errors within one hour: ", errCount)
				okCount = 0
				errCount = 0
				summCount = 0
			}
			time.Sleep(3 * time.Second)
			RabbitMQConnectionMetric.Set(exitCode)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":80", nil)
}

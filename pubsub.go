package main

import (
"context"
"encoding/json"
"log"
"net/http"
"regexp"
"strings"
"sync"

    "cloud.google.com/go/pubsub"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"

)

// Payload receveid from pub/sub
type Payload struct {
Temperature float64 `json:"temperature"`
Moisture float64 `json:"moisture"`
Ph float64 `json:"ph"`
DeviceID string `json:"deviceID"`
}

// Gauge metric temperature for Prometheus
var (
temperatureGauge = promauto.NewGauge(prometheus.GaugeOpts{
Name: "temperature_gauge",
Help: "Temperature in °C",
})
moistureGauge = promauto.NewGauge(prometheus.GaugeOpts{
Name: "moisture_gauge",
Help: "Moisture in %",
})
phGauge = promauto.NewGauge(prometheus.GaugeOpts{
Name: "ph_gauge",
Help: "ph in scale of 0-14",
})
idGauge = promauto.NewGauge(prometheus.GaugeOpts{
Name: "device_gauge",
Help: "device number",
})
)

// Pulling messages from pub/sub
func pullMessages() {
log.Println("STARTED PULLING MESSAGES")

    ctx := context.Background()

    // Set your $PROJECT_ID
    client, err := pubsub.NewClient(ctx, "ta-wsn")
    if err != nil {
    	log.Fatal(err)
    }

    // Set your $SUBSCRIPTION
    subID := "subs_data"
    var mu sync.Mutex

    sub := client.Subscription(subID)
    cctx, cancel := context.WithCancel(ctx)
    err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
    	mu.Lock()
    	defer mu.Unlock()

    	log.Print("Got message: " + string(msg.Data))

    	if !containsInitPayload(string(msg.Data)) {
    		var t Payload

    		err := json.Unmarshal(msg.Data, &t)
    		if err != nil {
    			log.Fatal(err)
    		}
    		if t.DeviceID == "3" {

    			temperatureGauge.Set(t.Temperature)
    			moistureGauge.Set(t.Moisture)
    			phGauge.Set(t.Ph)
    		}
    	}

    	msg.Ack()

    })
    if err != nil {
    	cancel()
    	log.Fatal(err)
    }
    cancel()

}

func containsInitPayload(payload string) bool {
var regex, \_ = regexp.Compile(`^\{[^}]*\}$`)

    if strings.Contains(payload, "node1-connected") || !(regex.MatchString(payload)) {
    	return true
    }
    return false

}

func main() {
go pullMessages()

    log.Println("STARTED PROMETHEUS")

    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":2112", nil)

}

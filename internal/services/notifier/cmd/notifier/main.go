package main

import (
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/infrastructure/kafka/consumer"

	"github.com/fatih/color"
)

//go:embed web
var staticFS embed.FS

const topic = "pvz.events-log"

type stored struct {
	Time string          `json:"time"`
	Raw  json.RawMessage `json:"raw"`
}

var (
	mu  sync.RWMutex
	buf []stored
)

func main() {
	brokers := "localhost:9092"
	cg, err := consumer.NewConsumerGroup(
		consumer.Config{
			Brokers:     []string{brokers},
			Topic:       topic,
			GroupID:     "pvz-notifier",
			ResetOffset: "earliest",
			AutoCommit:  true,
		},
		consumer.GroupHandler{
			OnMessage: handle,
		})
	if err != nil {
		log.Fatalf("consumer: %v", err)
	}
	defer cg.Close()

	content, _ := fs.Sub(staticFS, "web")

	http.Handle("/", http.FileServer(http.FS(content)))
	http.HandleFunc("/events", serveJSON)

	go func() {
		log.Println("HTTP  : http://localhost:8888")
		log.Fatal(http.ListenAndServe(":8888", nil))
	}()

	log.Printf("Kafka : brokers=%s topic=%s", brokers, topic)
	if err := cg.Run(context.Background()); err != nil &&
		!errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func handle(_ context.Context, rec *kgo.Record) error {
	var ev models.OrderEvent
	if err := json.Unmarshal(rec.Value, &ev); err != nil {
		log.Printf("bad json: %v", err)
		return nil
	}

	shortFmt(ev)

	pretty, _ := json.MarshalIndent(ev, "", "  ")
	color.New(color.FgCyan).Printf("%s\n\n", pretty)

	mu.Lock()
	buf = append(buf, stored{
		Time: time.Now().Format("15:04:05"),
		Raw:  append([]byte{}, pretty...),
	})
	if len(buf) > 200 {
		buf = buf[1:]
	}
	mu.Unlock()
	return nil
}

func shortFmt(e models.OrderEvent) {
	switch e.EventType {
	case models.OrderAccepted:
		color.Green("✅ ACCEPTED order=%s courier=%d", e.Order.ID, e.Actor.ID)
	case models.OrderReturnedToCourier:
		color.Red("↩️  RETURN-TO-COURIER order=%s", e.Order.ID)
	case models.OrderIssued:
		color.Blue("🎉 ISSUED order=%s client=%s", e.Order.ID, e.Order.UserID)
	case models.OrderReturnedByClient:
		color.Yellow("🙁 RETURN-BY-CLIENT order=%s", e.Order.ID)
	}
}

func serveJSON(w http.ResponseWriter, _ *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(buf)
}

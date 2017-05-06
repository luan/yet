package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/vito/go-sse/sse"
)

func main() {
	http.HandleFunc("/", yet)
	http.HandleFunc("/events", events)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func getLevel() (int, error) {
	doc, err := goquery.NewDocument("https://secure.tibia.com/community/?subtopic=characters&name=Shugoi")
	if err != nil {
		return 0, err
	}

	levelStr := doc.Find("#characters > div.Border_2  table:nth-child(1) tr:nth-child(5) > td:nth-child(2)").First().Text()
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		return 0, err
	}

	return level, nil
}

func yet(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("<style>body { text-align: center; font-family: sans-serif; }</style>"))
	w.Write([]byte(`<h1 id="result"></h1>`))
	w.Write([]byte(`<script>
		var source = new EventSource("/events");
		source.addEventListener('event', function(event) {
			var text = "No.";
			if (parseInt(event.data) >= 300) {
				text = "Yes!!! She finally did it, who would've thought.";
			}
			document.getElementById("result").innerHTML = text;
		}, false);
	</script>`))
}

func events(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Connection", "keep-alive")

	w.WriteHeader(http.StatusOK)

	flusher := w.(http.Flusher)
	flusher.Flush()
	eventID := 0

	for {
		level, err := getLevel()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		fmt.Println(level)

		err = sse.Event{
			ID:   fmt.Sprintf("%d", eventID),
			Name: "event",
			Data: []byte(strconv.Itoa(level)),
		}.Write(w)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		flusher.Flush()
		time.Sleep(5 * time.Second)

		eventID++
	}
}

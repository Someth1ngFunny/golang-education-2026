package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
	"syscall"
)

type Film struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

func fetchFilm(ctx context.Context, id int, client *http.Client) (Film, error) {
	url := fmt.Sprintf("https://homeworksite.site/%d/info.0.json", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Film{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Film{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Film{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var res Film
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return Film{}, err
	}
	return res, nil
}

func main() {
	from := flag.Int("from", -1, "id первого фильма")
	to := flag.Int("to", -1, "id последнего фильма")
	workers := flag.Int("workers", 10, "количество воркеров в worker pool")
	timeout := flag.Int("timeout", 5, "таймаут (секунд) одного HTTP-запроса")
	flag.Parse()

	if *from < 0 || *to < 0 || *from > *to {
		fmt.Fprintln(os.Stderr, "error: 0 <= from <= to")
		os.Exit(1)
	}
	if *workers <= 0 {
		fmt.Fprintln(os.Stderr, "error: workers > 0")
		os.Exit(1)
	}
	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "error: timeout > 0")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var n int = *to - *from + 1
	var client = &http.Client{Timeout: time.Duration(*timeout) * time.Second}

	films := make([]Film, n)
	isOK := make([]bool, n)
	errors := make([]error, n)

	jobs := make(chan int)
	var wg sync.WaitGroup

	for range *workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filmId := range jobs {
				film, err := fetchFilm(ctx, filmId, client)
				if err == nil {
					isOK[filmId-*from] = true
					films[filmId-*from] = film
				} else {
					isOK[filmId-*from] = false
					errors[filmId-*from] = err
				}
			}
		}()
	}

produce:
	for filmId := *from; filmId <= *to; filmId++ {
		select {
		case jobs <- filmId:
		case <- ctx.Done():
			break produce
		}
	}
	close(jobs)
	wg.Wait()

	for i := range n {
		switch {
		case isOK[i]:
			fmt.Printf("%d — %s — %d — %s\n", films[i].Id, films[i].Title, films[i].Year, films[i].Director)
		case errors[i] != nil:
			fmt.Printf("Fetch to film with id \"%d\" failed with error \"%v\"\n", i+*from, errors[i])
		default:
			fmt.Printf("Fetch to film with id \"%d\" failed due to timeout\n", i+*from)
		}
	}
}

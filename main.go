package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/rqlite/gorqlite"
)

type Quote struct {
	ID     int
	Author string
	Text   string
	URL    string
}

func main() {
	listenAddress := flag.String("listen-address", ":4000", "This service listen address.")
	connURL := flag.String("rqlite-url", "http://rqlite:4001", "The rqlite URL.")

	flag.Parse()

	conn, err := openConnection(*connURL)
	if err != nil {
		slog.Error("failed to open the rqlite connection", "error", err)
		os.Exit(1)
	}

	err = createSchemaAndData(conn)
	if err != nil {
		slog.Error("failed to create the schema and data", "error", err)
		os.Exit(1)
	}

	if *listenAddress == "" {
		quote, err := getQuote(conn)
		if err != nil {
			slog.Error("failed to get quote", "error", err)
			return
		}
		slog.Info("quote", "quote", quote)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dump, _ := httputil.DumpRequest(r, false)
		slog.Info("client request", "request", dump)
		quote, err := getQuote(conn)
		if err != nil {
			slog.Error("failed to get quote", "error", err)
			os.Exit(1)
		}
		body := fmt.Sprintf("%s\n", quote)
		w.Write([]byte(body))
	})
	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		slog.Error("failed to listen for http requests", "error", err)
		os.Exit(1)
	}
}

func openConnection(connURL string) (*gorqlite.Connection, error) {
	for {
		conn, err := gorqlite.Open(connURL)
		if err == nil {
			qr, err := conn.QueryOne("select sqlite_version()")
			if err == nil {
				var sqliteVersion string
				for qr.Next() {
					err = qr.Scan(&sqliteVersion)
					if err != nil {
						conn.Close()
						return nil, fmt.Errorf("failed to scan result: %w", err)
					}
				}
				slog.Info("connected to rqlite", "sqlite_version", sqliteVersion)
				return conn, nil
			}
		}
		if conn != nil {
			conn.Close()
		}
		time.Sleep(3 * time.Second)
	}
}

func createSchemaAndData(conn *gorqlite.Connection) error {
	_, err := conn.WriteOne(`
create table if not exists quote (
	id		integer	not null	primary key,
	author	text	not null,
	text	text	not null,
	url		text	null
) strict
`)
	if err != nil {
		return fmt.Errorf("failed to create the quote table: %w", err)
	}

	quotes := []Quote{
		{
			ID:     1,
			Author: "Homer Simpson",
			Text:   "To alcohol! The cause of... and solution to... all of life's problems.",
			URL:    "https://en.wikipedia.org/wiki/Homer_vs._the_Eighteenth_Amendment",
		},
		{
			ID:     2,
			Author: "President Skroob, Spaceballs",
			Text:   "You got to help me. I don't know what to do. I can't make decisions. I'm a president!",
			URL:    "https://en.wikipedia.org/wiki/Spaceballs",
		},
		{
			ID:     3,
			Author: "Pravin Lal",
			Text:   "Beware of he who would deny you access to information, for in his heart he dreams himself your master.",
			URL:    "https://alphacentauri.gamepedia.com/Peacekeeping_Forces",
		},
		{
			ID:     4,
			Author: "Edsger W. Dijkstra",
			Text:   "About the use of language: it is impossible to sharpen a pencil with a blunt axe. It is equally vain to try to do it with ten blunt axes instead.",
			URL:    "https://www.cs.utexas.edu/users/EWD/transcriptions/EWD04xx/EWD498.html",
		},
		{
			ID:     5,
			Author: "Gina Sipley",
			Text:   "Those hours of practice, and failure, are a necessary part of the learning process.",
		},
		{
			ID:     6,
			Author: "Henry Petroski",
			Text:   "Engineering is achieving function while avoiding failure.",
		},
		{
			ID:     7,
			Author: "Jen Heemstra",
			Text:   "Leadership is defined by what you do, not what you're called.",
			URL:    "https://twitter.com/jenheemstra/status/1260186699021287424",
		},
		{
			ID:     8,
			Author: "Ludwig van Beethoven",
			Text:   "Don't only practice your art, but force your way into its secrets; art deserves that, for it and knowledge can raise man to the Divine.",
		},
	}

	for _, quote := range quotes {
		err = insertQuote(conn, &quote)
		if err != nil {
			return fmt.Errorf("failed to insert quote: %w", err)
		}
	}

	return nil
}

func insertQuote(conn *gorqlite.Connection, quote *Quote) error {
	_, err := conn.WriteOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     "insert or ignore into quote(id, author, text, url) values(?, ?, ?, ?)",
			Arguments: []interface{}{quote.ID, quote.Author, quote.Text, quote.URL},
		})
	return err
}

func getQuote(conn *gorqlite.Connection) (string, error) {
	qr, err := conn.QueryOne(`
select text || ' -- ' || author as quote
from quote
order by random()
limit 1
`)
	if err != nil {
		return "", fmt.Errorf("failed to query: %w", err)
	}
	var quote string
	for qr.Next() {
		err = qr.Scan(&quote)
		if err != nil {
			return "", fmt.Errorf("failed to scan result: %w", err)
		}
	}
	return quote, nil
}

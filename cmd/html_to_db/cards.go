package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"RoyalWikiOverlay/domain"
	"RoyalWikiOverlay/infrastructure/sqlite"
)

const (
	dbPath  = "data/royalwiki.db"
	htmlDir = "wiki_html_cards"
)

func main() {
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatal(err)
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := sqlite.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	log.Println("🔍 Parsing HTML cards...")
	if err := parseAndSaveCards(db, htmlDir); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Cards import completed")
}

func parseAndSaveCards(db *sql.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO items (name, type, price, wiki_url)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	inserted := 0

	for _, file := range files {
		log.Printf("📄 Processing %s", filepath.Base(file))

		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		doc, err := html.Parse(strings.NewReader(string(content)))
		if err != nil {
			log.Printf("html parse error: %v", err)
			continue
		}

		items := extractCards(doc)

		for _, item := range items {
			if _, err := stmt.Exec(
				item.Name,
				item.Type,
				item.Price,
				item.WikiURL,
			); err != nil {
				log.Printf("insert error (%s): %v", item.Name, err)
				continue
			}
			inserted++
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("🧾 Cards processed: %d", inserted)
	return nil
}

// extractCards извлекает карточки из HTML-документа
func extractCards(doc *html.Node) []domain.Item {
	var items []domain.Item

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			var href, title string
			for _, a := range n.Attr {
				switch a.Key {
				case "href":
					href = a.Val
				case "title":
					title = a.Val
				}
			}

			if isCard(title) {
				items = append(items, domain.Item{
					Name:    normalizeCardName(title),
					Type:    domain.ItemTypeCard,
					Price:   0,
					WikiURL: href,
				})
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return items
}

func isCard(title string) bool {
	return strings.Contains(title, "Карта")
}

func normalizeCardName(title string) string {
	name := strings.TrimSpace(title)
	name = strings.TrimPrefix(name, "Карта ")
	return strings.Join(strings.Fields(name), " ")
}

package main

import (
	"encoding/csv"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/otiai10/gosseract/v2"
)

const (
	outputCSV  = "market_scan.csv"
	fps        = 4
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

type MarketRow struct {
	Item     string
	Quantity int
	Price    int
	Merchant string
	Date     string
	Time     string
}

func main() {
	fmt.Println("🖱️ Кликни ПЕРВУЮ точку области рынка")
	x1, y1 := robotgo.GetMousePos()
	robotgo.MouseClick("left")

	time.Sleep(2 * time.Second)

	fmt.Println("🖱️ Кликни ВТОРУЮ точку области рынка")
	x2, y2 := robotgo.GetMousePos()
	robotgo.MouseClick("left")

	rect := normalizeRect(x1, y1, x2, y2)
	fmt.Printf("📐 Область: %+v\n", rect)

	file, err := os.OpenFile(outputCSV, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	existing := loadExisting(file)

	client := gosseract.NewClient()
	defer client.Close()

	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	fmt.Println("📸 Сканирование рынка запущено... (Ctrl+C для выхода)")

	for range ticker.C {
		img, _ := robotgo.CaptureImg(rect)
		if img == nil {
			continue
		}

		client.SetImageFromBytes(imageToBytes(img))
		text, err := client.Text()
		if err != nil {
			continue
		}

		rows := parseMarketText(text)
		now := time.Now()

		for _, r := range rows {
			r.Date = now.Format(dateLayout)
			r.Time = now.Format(timeLayout)

			key := dedupeKey(r)
			if existing[key] {
				continue
			}

			writer.Write([]string{
				r.Item,
				strconv.Itoa(r.Quantity),
				strconv.Itoa(r.Price),
				r.Merchant,
				r.Date,
				r.Time,
			})
			existing[key] = true
			writer.Flush()

			log.Printf("💰 %s x%d = %d (%s)", r.Item, r.Quantity, r.Price, r.Merchant)
		}
	}
}

func normalizeRect(x1, y1, x2, y2 int) image.Rectangle {
	minX := min(x1, x2)
	minY := min(y1, y2)
	maxX := max(x1, x2)
	maxY := max(y1, y2)
	return image.Rect(minX, minY, maxX, maxY)
}

func imageToBytes(img image.Image) []byte {
	buf := robotgo.ToBitmapBytes(img)
	return buf
}

func parseMarketText(text string) []MarketRow {
	lines := strings.Split(text, "\n")
	var rows []MarketRow

	for _, l := range lines {
		// ОЧЕНЬ грубый MVP-парсер
		// Fire Card x10 150000 TraderJohn
		parts := strings.Fields(l)
		if len(parts) < 5 {
			continue
		}

		qty, _ := strconv.Atoi(parts[len(parts)-4][1:])
		price, _ := strconv.Atoi(parts[len(parts)-3])
		merchant := parts[len(parts)-1]
		item := strings.Join(parts[:len(parts)-4], " ")

		rows = append(rows, MarketRow{
			Item:     item,
			Quantity: qty,
			Price:    price,
			Merchant: merchant,
		})
	}

	return rows
}

func dedupeKey(r MarketRow) string {
	return fmt.Sprintf("%s|%d|%d|%s",
		r.Item,
		r.Quantity,
		r.Price,
		r.Merchant,
	)
}

func loadExisting(f *os.File) map[string]bool {
	f.Seek(0, 0)
	r := csv.NewReader(f)
	rows, _ := r.ReadAll()

	m := make(map[string]bool)
	for _, row := range rows {
		if len(row) < 6 {
			continue
		}
		key := fmt.Sprintf("%s|%s|%s|%s",
			row[0],
			row[1],
			row[2],
			row[3],
		)
		m[key] = true
	}
	return m
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

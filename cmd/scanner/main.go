package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/otiai10/gosseract/v2"
	hook "github.com/robotn/gohook"
)

const (
	outputCSV  = "data/market_scan.csv"
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

// TODO Хуета ёбанная не может блять нормально распознавать сука
func main() {
	hook.AddEvent("mleft")
	evChan := hook.Start()
	defer hook.End()

	var x1, y1, x2, y2 int

	fmt.Println("🖱️ Кликни ПЕРВУЮ точку области рынка")
	for ev := range evChan {
		if ev.Kind == hook.MouseDown && ev.Button == hook.MouseMap["left"] {
			x1, y1 = int(ev.X), int(ev.Y)
			break
		}
	}

	fmt.Println("🖱️ Кликни ВТОРУЮ точку области рынка")
	for ev := range evChan {
		if ev.Kind == hook.MouseDown && ev.Button == hook.MouseMap["left"] {
			x2, y2 = int(ev.X), int(ev.Y)
			break
		}
	}

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
	client.SetTessdataPrefix("C:\\Program Files\\Tesseract-OCR\\tessdata")
	client.Languages = []string{"eng", "rus"}
	defer client.Close()

	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	fmt.Println("📸 Сканирование рынка запущено... (Ctrl+C для выхода)")

	for range ticker.C {
		img, _ := robotgo.CaptureImg(rect.Min.X, rect.Min.Y, rect.Max.X-rect.Min.X, rect.Max.Y-rect.Min.Y)
		if img == nil {
			fmt.Println("Failed to capture image")
			continue
		}

		fmt.Println("Captured image successfully")

		// Save for debugging
		robotgo.Save(img, "debug_capture.jpg")

		err := client.SetImageFromBytes(imageToBytes(img))
		if err != nil {
			fmt.Println("Error setting image:", err)
			return
		}

		text, err := client.Text()
		if err != nil {
			fmt.Println("Error getting text:", err)
			continue
		}

		fmt.Println("OCR Text:", text)

		rows := parseMarketText(text)
		fmt.Println("Parsed rows:", len(rows))

		now := time.Now()

		for _, r := range rows {
			r.Date = now.Format(dateLayout)
			r.Time = now.Format(timeLayout)

			key := dedupeKey(r)
			if existing[key] {
				continue
			}

			err := writer.Write([]string{
				r.Item,
				strconv.Itoa(r.Quantity),
				strconv.Itoa(r.Price),
				r.Merchant,
				r.Date,
				r.Time,
			})
			if err != nil {
				return
			}
			existing[key] = true
			writer.Flush()

			log.Printf("💰 %s x%d = %d (%s)", r.Item, r.Quantity, r.Price, r.Merchant)
		}
	}
}

func normalizeRect(x1, y1, x2, y2 int) image.Rectangle {
	minX := getMin(x1, x2)
	minY := getMin(y1, y2)
	maxX := getMax(x1, x2)
	maxY := getMax(y1, y2)
	return image.Rect(minX, minY, maxX, maxY)
}

func imageToBytes(img image.Image) []byte {
	buf := new(bytes.Buffer)
	jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
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
	_, err := f.Seek(0, 0)
	if err != nil {
		return nil
	}
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

func getMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

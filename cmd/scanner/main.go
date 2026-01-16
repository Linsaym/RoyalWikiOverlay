package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/otiai10/gosseract/v2"
	hook "github.com/robotn/gohook"
	"golang.org/x/image/draw"
)

const (
	outputCSV  = "data/market_scan.csv"
	fps        = 1 // Еще меньше FPS для стабильности
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
	// Проверяем наличие Tesseract
	if _, err := os.Stat("C:\\Program Files\\Tesseract-OCR\\tessdata"); os.IsNotExist(err) {
		fmt.Println("❌ Tesseract не установлен или путь неверный")
		fmt.Println("Скачайте: https://github.com/UB-Mannheim/tesseract/wiki")
		return
	}

	hook.AddEvent("mleft")
	evChan := hook.Start()
	defer hook.End()

	fmt.Println("🖱️ Кликни ЛЕВЫЙ ВЕРХНИЙ угол таблицы")
	x1, y1 := waitForClick(evChan)
	fmt.Println("🖱️ Кликни ПРАВЫЙ НИЖНИЙ угол таблицы")
	x2, y2 := waitForClick(evChan)

	rect := normalizeRect(x1, y1, x2, y2)
	fmt.Printf("📐 Область: %d,%d - %d,%d\n", rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)

	// Создаем папки
	os.MkdirAll("data", 0755)
	os.MkdirAll("debug", 0755)

	file, err := os.OpenFile(outputCSV, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Заголовок CSV если файл новый
	stat, _ := file.Stat()
	if stat.Size() == 0 {
		writer.Write([]string{"Item", "Quantity", "Price", "Merchant", "Date", "Time"})
		writer.Flush()
	}

	existing := loadExisting(file)

	client := gosseract.NewClient()
	defer client.Close()

	// Ключевые настройки для русского
	client.SetTessdataPrefix("C:\\Program Files\\Tesseract-OCR\\tessdata")
	client.Languages = []string{"rus+eng"}
	client.SetPageSegMode(gosseract.PSM_SINGLE_BLOCK_VERT_TEXT) // PSM 5 - вертикальная ориентация
	client.SetVariable("preserve_interword_spaces", "1")
	// УБИРАЕМ whitelist чтобы не мешал
	// client.SetVariable("tessedit_char_whitelist", "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzАБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯабвгдеёжзийклмнопрстуфхцчшщъыьэюя.,-: ")

	// Вместо whitelist лучше использовать blacklist для исключения символов
	client.SetVariable("tessedit_char_blacklist", "!@#$%^&*()_+=|\\/[]{}<>`~")

	fmt.Println("✅ Настройки Tesseract применены")
	fmt.Println("📸 Сканирование запущено... (Ctrl+C для выхода)")

	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	scanCount := 0
	for range ticker.C {
		scanCount++
		fmt.Printf("\n=== Сканирование #%d ===\n", scanCount)

		// Захват экрана
		img, err := robotgo.CaptureImg(rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy())
		if err != nil {
			fmt.Println("❌ Ошибка захвата:", err)
			continue
		}

		// Предобработка изображения - ДРУГОЙ ПОДХОД
		processed := preprocessForTable(img)

		// OCR
		err = client.SetImageFromBytes(imageToBytes(processed))
		if err != nil {
			fmt.Println("❌ Ошибка установки изображения:", err)
			continue
		}

		text, err := client.Text()
		if err != nil {
			fmt.Println("❌ Ошибка OCR:", err)
			continue
		}

		// ИСПРАВЛЯЕМ OCR ошибки ПРАВИЛЬНО
		text = fixOCRCommonErrors(text)
		fmt.Println("📝 Распознанный текст:")
		fmt.Println(text)

		// Парсинг с учетом исправленных ошибок
		rows := parseMarketTextImproved(text)
		fmt.Printf("📊 Найдено строк: %d\n", len(rows))

		now := time.Now()
		newCount := 0

		for _, r := range rows {
			if r.Item == "" || r.Merchant == "" {
				continue
			}

			r.Date = now.Format(dateLayout)
			r.Time = now.Format(timeLayout)

			key := dedupeKey(r)
			if existing[key] {
				fmt.Printf("⚠️  Дубликат: %s\n", r.Item)
				continue
			}

			err := writer.Write([]string{
				strings.TrimSpace(r.Item),
				strconv.Itoa(r.Quantity),
				strconv.Itoa(r.Price),
				strings.TrimSpace(r.Merchant),
				r.Date,
				r.Time,
			})
			if err != nil {
				fmt.Println("❌ Ошибка записи:", err)
				continue
			}

			existing[key] = true
			newCount++

			fmt.Printf("✅ Добавлено: %s x%d = %d (%s)\n",
				r.Item, r.Quantity, r.Price, r.Merchant)
		}

		writer.Flush()
		if newCount > 0 {
			fmt.Printf("💾 Сохранено новых записей: %d\n", newCount)
		} else if len(rows) > 0 {
			fmt.Println("ℹ️  Все строки уже есть в базе")
		}
	}
}

func preprocessForTable(img image.Image) image.Image {
	// 1. Увеличиваем в 3 раза (для мелкого текста)
	bounds := img.Bounds()
	scaled := image.NewRGBA(image.Rect(0, 0, bounds.Dx()*3, bounds.Dy()*3))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), img, bounds, draw.Over, nil)

	// 2. Бинаризация (черно-белое с порогом)
	result := image.NewGray(scaled.Bounds())

	// Автоматическое определение порога
	for y := scaled.Bounds().Min.Y; y < scaled.Bounds().Max.Y; y++ {
		for x := scaled.Bounds().Min.X; x < scaled.Bounds().Max.X; x++ {
			r, g, b, _ := scaled.At(x, y).RGBA()
			// Преобразование в grayscale
			gray := (r*299 + g*587 + b*114) / 1000

			// Бинаризация с порогом 18000 (примерно 70 из 255)
			if gray > 18000 {
				result.SetGray(x, y, color.Gray{Y: 255}) // Белый
			} else {
				result.SetGray(x, y, color.Gray{Y: 0}) // Черный
			}
		}
	}

	return result
}

func fixOCRCommonErrors(text string) string {
	// Сначала исправляем цифры на буквы (основная проблема)
	replacements := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		// Цифры -> Русские буквы
		{regexp.MustCompile(`\b3([а-яА-Я])`), "З$1"}, // 3олотая -> Золотая
		{regexp.MustCompile(`\b8([а-яА-Я])`), "В$1"}, // 8оителя -> Воителя
		{regexp.MustCompile(`0([а-яА-Я])`), "О$1"},   // мага3ин -> магазин
		{regexp.MustCompile(`4([а-яА-Я])`), "Ч$1"},   // 4ен -> Чен
		{regexp.MustCompile(`6([а-яА-Я])`), "Б$1"},   // 6алница -> Балница

		// Общие паттерны из вашего примера
		{regexp.MustCompile(`мага3им\b`), "магазин"},
		{regexp.MustCompile(`3олотая`), "Золотая"},
		{regexp.MustCompile(`8оителя`), "Воителя"},
		{regexp.MustCompile(`4ен\.`), "Чен."},
		{regexp.MustCompile(`8алница`), "Балница"},
		{regexp.MustCompile(`мараге`), "магазин"}, // mar age исправление

		// Убираем лишние пробелы внутри чисел
		{regexp.MustCompile(`(\d)\s+(\d)`), "$1$2"}, // 15 350 -> 15350
	}

	for _, r := range replacements {
		text = r.pattern.ReplaceAllString(text, r.replacement)
	}

	// Убираем множественные пробелы
	reSpaces := regexp.MustCompile(`\s+`)
	text = reSpaces.ReplaceAllString(text, " ")

	// Исправляем слипшиеся слова
	reWords := regexp.MustCompile(`([а-яА-Я])(\d+)`)
	text = reWords.ReplaceAllString(text, "$1 $2")

	return strings.TrimSpace(text)
}

func parseMarketTextImproved(text string) []MarketRow {
	var rows []MarketRow

	// Разделяем на строки
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 10 { // Минимальная длина строки
			continue
		}

		// Пробуем несколько паттернов

		// Паттерн 1: "Золотая карта Воителя 15350 325 магазин"
		// Ищем последние два числа как количество и цену
		re := regexp.MustCompile(`^(.*?)\s+(\d+)\s+(\d+)\s+([^\d]+)$`)
		if matches := re.FindStringSubmatch(line); matches != nil {
			if len(matches) >= 5 {
				price, err1 := strconv.Atoi(matches[3])
				qty, err2 := strconv.Atoi(matches[2])

				// Проверяем что числа разумные
				if err1 == nil && err2 == nil && price > 0 && price < 100000000 && qty > 0 && qty < 10000 {
					rows = append(rows, MarketRow{
						Item:     strings.TrimSpace(matches[1]),
						Quantity: qty,
						Price:    price,
						Merchant: strings.TrimSpace(matches[4]),
					})
					continue
				}
			}
		}

		// Паттерн 2: "Золотая карта Воителя x10 150000 магазин"
		re2 := regexp.MustCompile(`^(.*?)\s+x?(\d+)\s+(\d+)\s+([^\d]+)$`)
		if matches := re2.FindStringSubmatch(line); matches != nil {
			if len(matches) >= 5 {
				price, err1 := strconv.Atoi(matches[3])
				qty, err2 := strconv.Atoi(matches[2])

				if err1 == nil && err2 == nil && price > 0 && qty > 0 {
					rows = append(rows, MarketRow{
						Item:     strings.TrimSpace(matches[1]),
						Quantity: qty,
						Price:    price,
						Merchant: strings.TrimSpace(matches[4]),
					})
					continue
				}
			}
		}

		// Паттерн 3: Более гибкий - ищем любые два числа в конце
		re3 := regexp.MustCompile(`^(.*?)\s+(\d+)\s+(\d+)\s*(.*?)$`)
		if matches := re3.FindStringSubmatch(line); matches != nil {
			if len(matches) >= 5 {
				// Пробуем разные комбинации
				num1, err1 := strconv.Atoi(matches[2])
				num2, err2 := strconv.Atoi(matches[3])
				merchant := strings.TrimSpace(matches[4])

				if err1 == nil && err2 == nil {
					// Определяем что есть что: обычно цена больше количества
					var qty, price int
					if num1 < 1000 && num2 > 1000 { // num1 - количество, num2 - цена
						qty, price = num1, num2
					} else if num2 < 1000 && num1 > 1000 { // наоборот
						qty, price = num2, num1
					} else { // берем по порядку
						qty, price = num1, num2
					}

					if price > 0 && qty > 0 {
						rows = append(rows, MarketRow{
							Item:     strings.TrimSpace(matches[1]),
							Quantity: qty,
							Price:    price,
							Merchant: merchant,
						})
					}
				}
			}
		}
	}

	return rows
}

func waitForClick(evChan chan hook.Event) (int, int) {
	for ev := range evChan {
		if ev.Kind == hook.MouseDown && ev.Button == hook.MouseMap["left"] {
			fmt.Printf("📍 Координаты: %d, %d\n", int(ev.X), int(ev.Y))
			return int(ev.X), int(ev.Y)
		}
	}
	return 0, 0
}

func normalizeRect(x1, y1, x2, y2 int) image.Rectangle {
	return image.Rect(
		min(x1, x2),
		min(y1, y2),
		max(x1, x2),
		max(y1, y2),
	)
}

func imageToBytes(img image.Image) []byte {
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 95})
	if err != nil {
		fmt.Println("❌ Ошибка кодирования:", err)
	}
	return buf.Bytes()
}

func dedupeKey(r MarketRow) string {
	return fmt.Sprintf("%s|%d|%d|%s",
		strings.ToLower(r.Item),
		r.Quantity,
		r.Price,
		strings.ToLower(r.Merchant),
	)
}

func loadExisting(f *os.File) map[string]bool {
	m := make(map[string]bool)

	f.Seek(0, 0)
	reader := csv.NewReader(f)

	records, err := reader.ReadAll()
	if err != nil || len(records) <= 1 {
		return m
	}

	for i, row := range records {
		if i == 0 || len(row) < 4 {
			continue
		}
		key := fmt.Sprintf("%s|%s|%s|%s",
			strings.ToLower(strings.TrimSpace(row[0])),
			strings.TrimSpace(row[1]),
			strings.TrimSpace(row[2]),
			strings.ToLower(strings.TrimSpace(row[3])),
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

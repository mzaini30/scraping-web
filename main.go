package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// Baca isi file kedalaman.txt
	depthStr, err := ioutil.ReadFile("kedalaman.txt")
	if err != nil {
		fmt.Println("Error reading depth file:", err)
		return
	}

	// Konversi isi file ke dalam bentuk bilangan bulat
	depth, err := strconv.Atoi(string(depthStr))
	if err != nil {
		fmt.Println("Error converting depth to integer:", err)
		return
	}

	// Lakukan scraping rekursif untuk setiap URL di file link.txt
	err = scrapeRecursively("link.txt", depth)
	if err != nil {
		fmt.Println("Error in recursive scraping:", err)
	}
}

func scrapeRecursively(filename string, iterations int) error {
	// Baca URL awal dari file
	urls, err := readURLsFromFile(filename)
	if err != nil {
		return err
	}

	// Lakukan scraping untuk setiap URL
	for _, url := range urls {
		// Lakukan scraping dan simpan hasil
		err := scrapeAndSaveLinks(url)
		if err != nil {
			fmt.Printf("Error scraping %s: %v\n", url, err)
			continue
		}

		// Lakukan scraping rekursif untuk hasil yang baru saja disimpan
		for i := 0; i < iterations; i++ {
			err := scrapeAndSaveLinksRecursively(url, i+1)
			if err != nil {
				fmt.Printf("Error in recursive scraping for %s (iteration %d): %v\n", url, i+1, err)
				continue
			}
		}
	}

	return nil
}

func scrapeAndSaveLinksRecursively(url string, iteration int) error {
	// Membersihkan URL agar dapat digunakan sebagai nama file
	cleanedURL := strings.ReplaceAll(url, "https://", "")
	cleanedURL = strings.ReplaceAll(cleanedURL, "http://", "")
	cleanedURL = strings.ReplaceAll(cleanedURL, "www.", "")
	cleanedURL = strings.ReplaceAll(cleanedURL, "/", "_")
	cleanedURL = strings.ReplaceAll(cleanedURL, ":", "_")

	// Membaca file hasil yang sudah ada
	filePath := fmt.Sprintf("hasil/%s.txt", cleanedURL)
	existingLinks, err := readURLsFromFile(filePath)
	if err != nil {
		return err
	}

	// Lakukan scraping untuk setiap link di file hasil
	for _, link := range existingLinks {
		err := scrapeAndSaveLinks(link)
		if err != nil {
			fmt.Printf("Error scraping %s: %v\n", link, err)
			continue
		}
	}

	// Menampilkan pesan untuk iterasi tertentu
	fmt.Printf("Links on %s (iteration %d) saved to %s\n", url, iteration, filePath)

	return nil
}

func readURLsFromFile(filename string) ([]string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Membersihkan karakter \r pada setiap baris
	content = bytes.Replace(content, []byte("\r"), []byte{}, -1)

	urls := strings.Split(string(content), "\n")
	return urls, nil
}

func scrapeAndSaveLinks(url string) error {
	// Membersihkan URL agar dapat digunakan sebagai nama file
	cleanedURL := strings.ReplaceAll(url, "https://", "")
	cleanedURL = strings.ReplaceAll(cleanedURL, "http://", "")
	cleanedURL = strings.ReplaceAll(cleanedURL, "www.", "")
	cleanedURL = strings.ReplaceAll(cleanedURL, "/", "_")
	cleanedURL = strings.ReplaceAll(cleanedURL, ":", "_")

	// Mendapatkan halaman HTML dari URL
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Membuat dokumen goquery dari HTML
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return err
	}

	// Mendapatkan base URL
	baseURL := cleanedURL

	// Jika baseURL tidak diakhiri dengan underscore, tambahkan underscore
	if !strings.HasSuffix(baseURL, "_") {
		baseURL += "_"
	}

	// Mendapatkan semua link pada halaman saat ini
	var links []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")

		// Menambahkan prefix jika link dimulai dengan karakter "/"
		if strings.HasPrefix(link, "/") {
			// Jika baseURL berisi karakter "_", kita ambil bagian sebelumnya
			if strings.Contains(baseURL, "_") {
				parts := strings.Split(baseURL, "_")
				baseURL = parts[0]
			}

			// Menambahkan https/http pada link
			link = "https://" + baseURL + link
			links = append(links, link)
		} else if strings.HasPrefix(link, "http") {
			links = append(links, link)
		}
	})

	// Menghapus duplikat link
	links = removeDuplicates(links)

	// Menghapus baris yang tidak memiliki isi
	links = removeEmptyLines(links)

	// Menyortir slice links
	sort.Strings(links)

	// Membuat folder "hasil" jika belum ada
	if _, err := os.Stat("hasil"); os.IsNotExist(err) {
		err := os.Mkdir("hasil", 0755)
		if err != nil {
			return err
		}
	}

	// Menyimpan link ke file
	outputFilePath := fmt.Sprintf("hasil/%s.txt", cleanedURL)
	err = saveLinksToFile(outputFilePath, links)
	if err != nil {
		return err
	}

	fmt.Printf("Links on %s saved to %s\n", url, outputFilePath)
	return nil
}

func saveLinksToFile(filename string, links []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Menyimpan link ke file
	for _, link := range links {
		_, err := file.WriteString(link + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func removeDuplicates(links []string) []string {
	uniqueLinks := make(map[string]struct{})
	var result []string
	for _, link := range links {
		uniqueLinks[link] = struct{}{}
	}
	for link := range uniqueLinks {
		result = append(result, link)
	}
	return result
}

func removeEmptyLines(links []string) []string {
	var result []string
	for _, link := range links {
		if link != "" {
			result = append(result, link)
		}
	}
	return result
}

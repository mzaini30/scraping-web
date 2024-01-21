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

	// Iterasi sebanyak yang ditentukan di kedalaman.txt
	for i := 1; i <= iterations; i++ {
		// Lakukan scraping untuk setiap URL
		for _, url := range urls {
			// Lakukan scraping dan simpan hasil
			err := scrapeAndSaveLinks(url, i)
			if err != nil {
				fmt.Printf("Error scraping %s: %v\n", url, err)
				continue
			}
		}
		// Ambil semua link dari hasil.txt (termasuk yang sudah diolah sebelumnya)
		// untuk digunakan pada iterasi berikutnya
		urls, err = readURLsFromFile("hasil.txt")
		if err != nil {
			return err
		}
	}

	return nil
}

func scrapeAndSaveLinks(url string, iteration int) error {
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
	// if _, err := os.Stat("hasil"); os.IsNotExist(err) {
	// 	err := os.Mkdir("hasil", 0755)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// Menyimpan link ke file hasil.txt dengan menambahkan hasil iterasi sebelumnya
	// outputFilePath := fmt.Sprintf("hasil/%s.txt", cleanedURL)
	err = saveLinksToFile("hasil.txt", links, iteration)
	if err != nil {
		return err
	}

	fmt.Printf("Links on %s (iteration %d) saved to hasil.txt\n", url, iteration)
	return nil
}

func saveLinksToFile(filename string, links []string, iteration int) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Menyimpan link ke file dengan menambahkan hasil iterasi sebelumnya
	for _, link := range links {
		_, err := file.WriteString(link + fmt.Sprintf(" (iteration %d)\n", iteration))
		if err != nil {
			return err
		}
	}

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

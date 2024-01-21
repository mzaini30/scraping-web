import scrapy
import os

class MySpider(scrapy.Spider):
    name = 'my_spider'
    start_urls = []  # Ganti dengan daftar URL awal dari file link.txt

    # Baca URL awal dari file link.txt
    with open('link.txt', 'r') as file:
        start_urls = [url.strip() for url in file.readlines()]

    def parse(self, response):
        # Mendapatkan semua link pada halaman saat ini
        links = response.css('a::attr(href)').extract()

        # Menyimpan link ke file
        output_file_path = f'hasil/{response.url.replace("https://", "").replace(".", "_").replace("/", "_")}.txt'
        with open(output_file_path, 'w') as output_file:
            output_file.write('\n'.join(links))

        # Cetak link untuk saat ini
        print(f'Links on {response.url} saved to {output_file_path}')

        # Menjalankan parsing pada setiap link untuk merayapi lebih lanjut
        for link in links:
            yield scrapy.Request(url=link, callback=self.parse)

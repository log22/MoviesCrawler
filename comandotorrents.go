package moviescrawler

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	ht "golang.org/x/net/html"
)

type ComandoTorrents struct{}

func (_ ComandoTorrents) SearchMovie(Term string, Page int) FoundMovies {
	defer recoveryPanic()
	headers := make(map[string]string)

	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36 OPR/68.0.3618.63"
	request := WebRequest{headers, nil}

	html := request.Get("https://comandotorrents.org/page/" + strconv.Itoa(Page) + "/?s=" + url.QueryEscape(Term))

	regex := regexp.MustCompile(`(?ms)<h2 class="entry-title" itemprop="headline"><a href="(.*?)">(.*?)<\/a><\/h2>.*?src="(.*?)"`)
	matches := regex.FindAllStringSubmatch(string(html), -1)
	var found FoundMovies
	for _, match := range matches {
		found = append(found, Movie{ht.UnescapeString(match[2]), match[3], match[1]})
	}
	return found
}

func (_ ComandoTorrents) GetDownloadLinks(Link string) FoundMagnetLinks {
	headers := make(map[string]string)
	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36 OPR/68.0.3618.63"
	html := WebRequest{headers, nil}.Get(Link)

	regex := regexp.MustCompile(`<p style="text-align: center;"><(?:strong|b)>(.*?)(?:<|\[|\|).*?\n.*href=".*?id=(.*?)&ref=(.*?)&titulo=(.*?)"`)
	matches := regex.FindAllStringSubmatch(string(html), -1)

	var foundLinks FoundMagnetLinks
	var wg sync.WaitGroup
	mutex := &sync.Mutex{}
	for _, match := range matches {
		
		if !strings.Contains(match[1], "reprodução") {
			wg.Add(1)
			go func(wg *sync.WaitGroup, foundLinks *FoundMagnetLinks, mutex *sync.Mutex) {
				defer wg.Done()
				magnetHeader := make(map[string]string)
				magnetHeader["Cookie"] = "nome=" + match[4] + "; ref=" + match[3] + "; idcriptografada=" + match[2]
				magnetHeader["Referer"] = "https://www.adssuper.com"

				magnetHtml := WebRequest{magnetHeader, nil}.Get("https://www.mastercuriosidadesbr.com/resenha-sobre-o-filme-malevola-2-dona-do-mal-parte-final/")
				magnetRegex := regexp.MustCompile(`(?m)href="(magnet:.+?)"`)
				magnetLink := magnetRegex.FindStringSubmatch(magnetHtml)[1]

				mutex.Lock()
				(*foundLinks) = append((*foundLinks), DownloadLink{Title: ht.UnescapeString(strings.Trim(match[1], " ")), MagnetLink: ht.UnescapeString(magnetLink)})
				mutex.Unlock()
			}(&wg, &foundLinks, mutex)
		}
	}

	wg.Wait()

	return foundLinks
}

func (_ ComandoTorrents) GetNumberOfPages(Term string) int {
	defer recoveryPanic()

	var headers map[string]string
	headers = make(map[string]string)

	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36 OPR/68.0.3618.63"
	request := WebRequest{headers, nil}

	html := request.Get("https://comandotorrents.org/?s=" + url.QueryEscape(Term))

	regex := regexp.MustCompile(`(?ms)<a class="last" href=".*?/([0-9]+)/`)
	result := regex.FindStringSubmatch(html)

	if result == nil {
		return 1
	}

	pageResult, _ := strconv.Atoi(result[1])

	if pageResult == 0 {
		return 1
	} else {
		return pageResult
	}
}

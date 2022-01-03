package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/semaphore"
)

var sem = semaphore.NewWeighted(5)
var pixivOrigin = "http://" + os.Getenv("PIXIV_API_HOST") + os.Getenv("PIXIV_API_PORT")

type PixivAccess struct {
}

func (a PixivAccess) CheckUrl(target url.URL) (string, string) {
	if target.Host == "i.pximg.net" && (strings.Contains(target.Path, "img-master") || strings.Contains(target.Path, "img-original") || strings.Contains(target.Path, "user-profile") || strings.Contains(target.Path, "background/img")) {
		lastPath := getLastPath(target.Path)
		id := strings.Split(lastPath, "_")[0]
		return "pic", id
	} else if target.Host == "www.pixiv.net" {
		if strings.Contains(target.Path, "artworks") {
			return "pic", getLastPath(target.Path)
		} else if strings.Contains(target.Path, "users") {
			return "author", getLastPath(target.Path)
		}
	}

	return "", ""
}

func getLastPath(path string) string {
	tokens := strings.Split(path, "/")
	return tokens[len(tokens)-1]
}

type pixivReturn struct {
	Illust illust `json:"illust"`
}

type illust struct {
	Caption    string `json:"caption"`
	CreateDate string `json:"create_date"`
	Height     int    `json:"height"`
	Id         int    `json:"id"`
	ImageUrls  struct {
		Large        string `json:"large"`
		Medium       string `json:"medium"`
		SquareMedium string `json:"square_medium"`
		Original     string `json:"Original"`
	} `json:"image_urls"`
	IsBookmarked bool `json:"is_bookmarked"`
	IsMuted      bool `json:"is_muted"`
	MetaPages    []struct {
		ImageUrls struct {
			Large        string `json:"large"`
			Medium       string `json:"medium"`
			SquareMedium string `json:"square_medium"`
			Original     string `json:"Original"`
		} `json:"image_urls"`
	} `json:"meta_pages"`
	MetaSinglePage struct {
		OriginalImageUrl string `json:"original_image_url"`
	} `json:"meta_single_page"`
	PageCount   int `json:"page_count"`
	Restrict    int `json:"restrict"`
	SanityLevel int `json:"sanity_level"`
	Series      struct {
		Id    int    `json:"id"`
		Title string `json:"title"`
	} `json:"series"`
	Tags []struct {
		Name           string `json:"name"`
		TranslatedName string `json:"translated_name"`
	} `json:"tags"`
	Title          string   `json:"title"`
	Tools          []string `json:"tools"`
	TotlaBookmarks int      `json:"total_bookmarks"`
	TotalComments  int      `json:"total_comments"`
	TotalView      int      `json:"total_view"`
	Type           string   `json:"type"`
	User           struct {
		Account          string `json:"account"`
		Id               int    `json:"id"`
		IsFollowed       bool   `json:"is_followed"`
		Name             string `json:"name"`
		ProfileImageUrls struct {
			Medium string `json:"medium"`
		} `json:"profile_image_urls"`
	} `json:"user"`
	Visible   bool `json:"visible"`
	Width     int  `json:"width"`
	XRestrict int  `json:"x_restrict"`
}

func (a PixivAccess) GetPicProfile(id string) (Profile, error) {
	res, err := http.Get(pixivOrigin + "/pic/" + id)
	if err != nil {
		return Profile{}, err
	}

	defer res.Body.Close()

	var ret pixivReturn
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		fmt.Println(err)
		return Profile{}, err
	}
	retIllust := ret.Illust

	return makeProfile(id, retIllust), nil
}

func makeProfile(id string, ilst illust) Profile {

	pageUrl := "https://www.pixiv.net/artworks/" + id
	var imageUrls []string
	if len(ilst.MetaPages) > 0 {
		for _, url := range ilst.MetaPages {
			imageUrls = append(imageUrls, url.ImageUrls.Original)
		}
	} else {
		imageUrls = []string{ilst.MetaSinglePage.OriginalImageUrl}
	}
	tags := []Tag{}
	for _, tag := range ilst.Tags {
		tags = append(tags, Tag{Name: tag.Name, Alt: []string{tag.TranslatedName}})
	}

	result := Profile{Id: strconv.Itoa(ilst.Id), Site: "pixiv", Title: ilst.Title, Description: ilst.Caption, Author: Account{Account: ilst.User.Account, Id: strconv.Itoa(ilst.User.Id), Name: ilst.User.Name, Follow: ilst.User.IsFollowed, ImageUrl: ilst.User.ProfileImageUrls.Medium}, Created: ilst.CreateDate, Size: Size{Width: ilst.Width, Height: ilst.Height}, PageUrl: pageUrl, ImageUrls: imageUrls, Bookmark: ilst.IsBookmarked, Tags: tags}

	return result
}

type picsReturn struct {
	Illusts []illust `json:"illusts"`
	NextUrl string   `json:"next_url"`
}

type indexedIllusts struct {
	Index   int
	Illusts []illust
}

const ILUSTS_PER_PAGE = 30

func (a PixivAccess) GetPicsOfAccount(id string, illustsCount int) (ProfilesTags, error) {
	pages := illustsCount/ILUSTS_PER_PAGE + 1

	sem := make(chan string, 4)
	resCh := make(chan indexedIllusts)
	errCh := make(chan error)
	defer close(sem)
	defer close(resCh)
	defer close(errCh)
	for i := 1; i <= pages; i++ {
		go getPicsOnePage(id, i, sem, resCh, errCh)
	}

	illustsIndexedList := make([]indexedIllusts, pages)
	for i := 1; i <= pages; i++ {
		select {
		case illustsIndexed := <-resCh:
			illustsIndexedList[illustsIndexed.Index-1] = illustsIndexed

		case err := <-errCh:
			return ProfilesTags{}, err
		}
	}

	var result ProfilesTags
	for _, ilstsIdx := range illustsIndexedList {
		for _, ilst := range ilstsIdx.Illusts {
			result.AddPicture(makeProfile(strconv.Itoa(ilst.Id), ilst))
		}
	}

	return result, nil
}

func getPicsOnePage(id string, page int, sem chan string, resCh chan indexedIllusts, errCh chan error) {
	url := fmt.Sprint(pixivOrigin + "/author/", id, "/pics?page=", page)
	sem <- url
	res, err := http.Get(url)
	if err != nil {
		errCh <- err
		<-sem
		return
	}

	defer res.Body.Close()

	var ret picsReturn
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		errCh <- err
		<-sem
		return
	}

	resCh <- indexedIllusts{Index: page, Illusts: ret.Illusts}
	<-sem
}

type pixivAuthorReturn struct {
	Profile struct {
		AddressId                  int    `json:"address_id"`
		BackgroundImageUrl         string `json:"background_image_url"`
		Birth                      string `json:"birth"`
		BirthDay                   string `json:"birth_day"`
		BirthYear                  int    `json:"birth_year"`
		CountryCode                string `json:"country_code"`
		Gender                     string `json:"gender"`
		IsPremium                  bool   `json:"is_premium"`
		IsUsingCustomProfileImage  bool   `json:"is_using_custom_profile_image"`
		Job                        string `json:"job"`
		JobId                      int    `json:"job_id"`
		PawooUrl                   string `json:"pawoo_url"`
		Region                     string `json:"region"`
		TotalFollowUsers           int    `json:"total_follow_users"`
		TotalIllustBookmarksPublic int    `json:"total_illust_bookmarks_public"`
		TotalIllustSeries          int    `json:"total_illust_series"`
		TotalIllusts               int    `json:"total_illusts"`
		TotalManga                 int    `json:"total_manga"`
		TotalMypixivUsers          int    `json:"total_mypixiv_users"`
		TotalMovelSeries           int    `json:"total_novel_series"`
		TotalNovels                int    `json:"total_novels"`
		TwitterAccount             string `json:"twitter_account"`
		TwitterUrl                 string `json:"twitter_url"`
		WebPage                    string `json:"webpage"`
	} `json:"profile"`
	ProfilePublicity struct {
		BirthDay  string `json:"birth_day"`
		BirthYear string `json:"birth_year"`
		Gender    string `json:"gender"`
		Job       string `json:"job"`
		Pawoo     bool   `json:"pawoo"`
		Region    string `json:"region"`
	} `json:"profile_publicity"`
	User struct {
		Account          string `json:"account"`
		Comment          string `son:"comment"`
		Id               int    `json:"id"`
		IsFollowed       bool   `json:"is_followed"`
		Name             string `json:"name"`
		ProfileImageUrls struct {
			Medium string `json:"medium"`
		} `json:"profile_image_urls"`
	} `json:"user"`
	Workspace struct {
		Chair             string `json:"chair"`
		Comment           string `json:"comment"`
		Desk              string `json:"desk"`
		Desktop           string `json:"desktop"`
		Montor            string `json:"monitor"`
		Mouse             string `json:"mouse"`
		Music             string `json:"music"`
		Pc                string `json:"pc"`
		Printer           string `json:"printer"`
		Scanner           string `json:"scanner"`
		Tablet            string `json:"tablet"`
		Tool              string `json:"tool"`
		WorkspaceImageUrl string `json:"workspace_image_url"`
	} `json:"workspace"`
}

type pixivUserProf struct {
	User map[string]struct {
		UserId   string `json:"userId"`
		Name     string `json:"name"`
		Image    string `json:"image"`
		ImageBig string `json:"imageBig"`
		WebPage  string `json:"webPage"`
		Social   map[string]struct {
			Url string `json:"url"`
		} `json:"social"`
	} `json:"user"`
}

func (a PixivAccess) GetAccountProfile(id string) (AccountProfile, error) {
	res, err := http.Get(pixivOrigin + "/author/" + id)
	if err != nil {
		return AccountProfile{}, err
	}

	defer res.Body.Close()

	var ret pixivAuthorReturn
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		fmt.Println(err)
		return AccountProfile{}, err
	}

	externals := []ExternalSite{}
	if ret.Profile.WebPage != "" {
		externals = append(externals, ExternalSite{Site: "webpage", Url: ret.Profile.WebPage})
	}
	if ret.Profile.TwitterUrl != "" {
		externals = append(externals, ExternalSite{Site: "twitter", Id: ret.Profile.TwitterAccount, Url: ret.Profile.TwitterUrl})
	}
	if ret.Profile.PawooUrl != "" {
		externals = append(externals, ExternalSite{Site: "pawoo", Url: ret.Profile.PawooUrl})
	}

	response, err := http.Get("https://www.pixiv.net/users/" + id)
	if err != nil {
		return AccountProfile{}, err
	}
	defer response.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(response.Body)
	doc.Find("meta").Each(func(_ int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		if name == "preload-data" {
			contentStr, _ := s.Attr("content")

			var content pixivUserProf
			json.Unmarshal([]byte(contentStr), &content)
			for key, value := range content.User[id].Social {
				exists := false
				for _, exs := range externals {
					exists = exists || (exs.Site == key)
				}
				if !exists {
					externals = append(externals, ExternalSite{Site: key, Url: value.Url})
				}
			}
		}
	})

	result := AccountProfile{Id: strconv.Itoa(ret.User.Id), Site: "pixiv", Account: ret.User.Account, Name: ret.User.Name, BackgroundUrl: ret.Profile.BackgroundImageUrl, ImageUrl: ret.User.ProfileImageUrls.Medium, Introduction: ret.User.Comment, IllustsCount: ret.Profile.TotalIllusts, PageUrl: "https://www.pixiv.net/users/" + id, ExternalSites: externals}

	return result, nil
}

func (a PixivAccess) Download(target url.URL) (DownloadResult, error) {
	fmt.Println(target.String())
	sem.Acquire(context.Background(), 1)
	defer sem.Release(1)
	response, err := http.Get(pixivOrigin + "/download?page=" + target.String())
	if err != nil {
		return DownloadResult{}, err
	}

	return DownloadResult{ContentType: response.Header.Get("Content-Type"), LastModified: response.Header.Get("Last-Modified"), Body: response.Body}, nil
}

package main

import (
  "io"
  "net/url"
)

type SiteAccess interface {
  CheckUrl(target url.URL) (string, string)
  GetPicProfile(id string) (Profile, error)
  GetAccountProfile(id string) (AccountProfile, error)
  GetPicsOfAccount(id string, illustsCount int) (ProfilesTags, error)
  Download(target url.URL) (DownloadResult, error)
}

type Profile struct {
  Id string `json:"id"`
  Site string `json:"site"`
  Title string `json:"title"`
  Description string `json:"description"`
  Author Account `json:"author"`
  Created string `json:"created"`
  Size Size `json:"size"`
  PageUrl string `json:"pageURL"`
  ImageUrls []string `json:"imageURLs"`
  Bookmark bool `json:" bookmark"`
  Tags []Tag `json:"tags"`
}

type Tag struct {
  Name string `json:"name"`
  Alt []string `json:"alt"`
}

type ProfilesTags struct {
  Pictures []Profile `json:"pictures"`
  Tags []tagCount `json:"tags"`
}

type tagCount struct {
  Tag Tag `json:"tag"`
  Count int `json:"count"`
}

func (pt *ProfilesTags) AddPicture(pic Profile) {
  pt.Pictures = append(pt.Pictures, pic);

  for _, tag := range pic.Tags {
    found := false
    for i := 0 ; i < len(pt.Tags) ; i++ {
      myTagCount := &pt.Tags[i]
      if tag.Name == myTagCount.Tag.Name {
        found = true
        myTagCount.Count += 1
        break
      }
    }
    if !found {
      pt.Tags = append(pt.Tags, tagCount{tag, 1})
    }
  }
}

type Illustrator struct {
  Name string `json:"name"`
  Accounts []Account `json:"accounts"`
}

type Account struct {
  Account string `json:"account"`
  Id string `json:"id"`
  Name string `json:"name"`
  Follow bool `json:"follow"`
  ImageUrl string `json:"imageURL"`
}

type Size struct {
  Width int `json:"width"`
  Height int `json:"height"`
}

type AccountProfile struct {
  Id string `json:"id"`
  Site string `json:"site"`
  Account string `json:"account"`
  Name string `json:"name"`
  ImageUrl string `json:"imageURL"`
  BackgroundUrl string `json:"backgroundURL"`
  Introduction string `json:"introduction"`
  IllustsCount int `json:"illustCount"`
  PageUrl string `json:"pageURL"`
  ExternalSites []ExternalSite `json:"externalSites"`
}

type ExternalSite struct {
  Site string `json:"site"`
  Id string `json:"id"`
  Url string `json:"URL"`
}

type DownloadResult struct {
  ContentType string
  LastModified string
  Body io.ReadCloser
}

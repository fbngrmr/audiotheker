package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hasura/go-graphql-client"
	"github.com/spf13/cobra"
)

var GRAPHQL_ENDPOINT = "https://api.ardaudiothek.de/graphql"

type QueryType int64

const (
	Episode QueryType = iota
	Collection
	Program
	Unknown
)

func init() {
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:   "download [URL] [targetDirectory]",
	Short: "Download all episodes of a program/collection or an individual episode in the ARD Audiothek.",
	Long:  "Download all episodes of a program/collection or an individual episode in the ARD Audiothek. Limited to 100 episodes.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		run(args[0], args[1])
	},
}

func downloadFile(url string, targetDirectory string) (err error) {
	fileName := path.Base(url)
	filePath := filepath.Join(targetDirectory, fileName)

	out, createFileErr := os.Create(filePath)
	if createFileErr != nil {
		return createFileErr
	}
	defer out.Close()

	httpResponse, httpErr := http.Get(url)
	if httpErr != nil {
		return httpErr
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad HTTP status: %s", httpResponse.Status)
	}

	_, copyError := io.Copy(out, httpResponse.Body)
	if copyError != nil {
		return copyError
	}

	return nil
}

func extractDownloadUrls(response *ItemsResponse) []string {
	var urls []string

	for _, nodes := range response.Result.Items.Nodes {
		for _, audios := range nodes.Audios {
			if audios.DownloadUrl != "" {
				urls = append(urls, audios.DownloadUrl)
			}
		}
	}

	return urls
}

func extractQueryId(url string) (string, QueryType) {
	normalizedUrl := strings.TrimSuffix(url, "/")

	// Program
	programPattern := regexp.MustCompile(`^https:\/\/www\.ardaudiothek\.de\/sendung\/.*\/(\d*)$`)
	matches := programPattern.FindStringSubmatch(normalizedUrl)

	if matches != nil {
		return matches[1], Program
	}

	// Collection
	collectionPattern := regexp.MustCompile(`^https:\/\/www\.ardaudiothek\.de\/sammlung\/.*\/(\d*)$`)
	matches = collectionPattern.FindStringSubmatch(normalizedUrl)

	if matches != nil {
		return matches[1], Collection
	}

	// Episode
	episodePattern := regexp.MustCompile(`^https:\/\/www\.ardaudiothek\.de\/episode\/.*\/(\d*)$`)
	matches = episodePattern.FindStringSubmatch(normalizedUrl)

	if matches != nil {
		return matches[1], Episode
	}

	return "", Unknown
}

type ItemsResponse struct {
	Result struct {
		Items struct {
			Nodes []struct {
				Audios []struct {
					DownloadUrl string
				}
			}
		}
	}
}

func sendGraphQlQuery(query string, variables map[string]interface{}, response interface{}) error {
	client := graphql.NewClient(GRAPHQL_ENDPOINT, nil)

	rawGraphqlResponse, graphQlErr := client.ExecRaw(context.Background(), query, variables)
	if graphQlErr != nil {
		return graphQlErr
	}

	if jsonError := json.Unmarshal(rawGraphqlResponse, &response); jsonError != nil {
		return jsonError
	}

	return nil
}

func getProgramUrls(queryId string) ([]string, error) {
	query := `query ProgramSetEpisodesQuery($id: ID!, $offset: Int!, $count: Int!) {
		result: programSet(id: $id) {
			items(
				offset: $offset
				first: $count
				orderBy: PUBLISH_DATE_DESC
				filter: { isPublished: { equalTo: true } }
			) {
				nodes {
					audios {
						downloadUrl
					}
				}
			}
		}
  	}`
	variables := map[string]interface{}{
		"id":     queryId,
		"offset": 0,
		"count":  100,
	}

	var response ItemsResponse
	graphQlError := sendGraphQlQuery(query, variables, &response)

	if graphQlError != nil {
		return nil, graphQlError
	}

	urls := extractDownloadUrls(&response)

	return urls, nil
}

func getCollectionUrls(queryId string) ([]string, error) {
	query := `query EpisodesQuery($id: ID!, $offset: Int!, $limit: Int!) {
		result: editorialCollection(id: $id, offset: $offset, limit: $limit) {
			items {
				nodes {
			  		id
			  		audios {
						url
						downloadUrl
						allowDownload
			  		}
				}
		  	}
		}
	}`
	variables := map[string]interface{}{
		"id":     queryId,
		"offset": 0,
		"limit":  100,
	}

	var response ItemsResponse
	graphQlError := sendGraphQlQuery(query, variables, &response)

	if graphQlError != nil {
		return nil, graphQlError
	}

	urls := extractDownloadUrls(&response)

	return urls, nil
}

type ItemResponse struct {
	Result struct {
		Audios []struct {
			DownloadUrl *string
		}
	}
}

func getEpisodeUrls(queryId string) ([]string, error) {
	query := `query EpisodeQuery($id: ID!) {
		result: item(id: $id) {
		  	audios {
				downloadUrl
		  	}
		}
	}`
	variables := map[string]interface{}{
		"id": queryId,
	}

	var response ItemResponse
	graphQlError := sendGraphQlQuery(query, variables, &response)

	if graphQlError != nil {
		return nil, graphQlError
	}

	var urls []string
	for _, audios := range response.Result.Audios {
		if audios.DownloadUrl != nil {
			if *audios.DownloadUrl != "" {
				urls = append(urls, *audios.DownloadUrl)
			}
		}
	}

	return urls, nil
}

func getDownloadUrls(url string) ([]string, error) {
	queryId, queryType := extractQueryId(url)

	switch queryType {
	case Episode:
		return getEpisodeUrls(queryId)

	case Collection:
		return getCollectionUrls(queryId)

	case Program:
		return getProgramUrls(queryId)

	default:
		return nil, fmt.Errorf("URL is not supported: %s", url)
	}
}

func run(url string, targetDirectory string) {
	urls, err := getDownloadUrls(url)

	if err != nil {
		panic(err)
	}

	for _, url := range urls {
		downloadErr := downloadFile(url, targetDirectory)

		if downloadErr != nil {
			fmt.Printf("Downloading file %s failed with error: %v\n", url, downloadErr)
		}
	}
}

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

var graphqlEndPoint = "https://api.ardaudiothek.de/graphql"

type QueryType int64

const (
	Episode QueryType = iota
	Collection
	Program
	Unknown
)

type File struct {
	url      string
	fileName string
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 äÄöÖüÜ\-\_]+`)

func toFileName(episodeTitle string, programTitle string) string {
	value := programTitle + "_-_" + episodeTitle
	return strings.Replace(nonAlphanumericRegex.ReplaceAllString(strings.Replace(value, "/", "_", -1), ""), " ", "_", -1)
}

func downloadFile(file File, targetDirectory string) (err error) {
	fileExtension := filepath.Ext(path.Base(file.url))
	filePath := filepath.Join(targetDirectory, file.fileName+fileExtension)

	out, createFileErr := os.Create(filePath)
	if createFileErr != nil {
		return createFileErr
	}
	defer out.Close()

	httpResponse, httpErr := http.Get(file.url)
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

func extractDownloadUrls(response *ItemsResponse) []File {
	var files []File

	for _, nodes := range response.Result.Items.Nodes {
		for _, audios := range nodes.Audios {
			if audios.DownloadUrl != "" {
				files = append(files, File{url: audios.DownloadUrl, fileName: toFileName(audios.Title, nodes.ProgramSet.Title)})
			} else if audios.Url != "" {
				files = append(files, File{url: audios.Url, fileName: toFileName(audios.Title, nodes.ProgramSet.Title)})
			}
		}
	}

	return files
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
				ProgramSet struct {
					Title string
				}
				Audios []struct {
					Title         string
					DownloadUrl   string
					AllowDownload bool
					Url           string
				}
			}
		}
	}
}

func getProgramUrls(queryId string) ([]File, error) {
	query := `query ProgramSetEpisodesQuery($id: ID!, $offset: Int!, $count: Int!) {
		result: programSet(id: $id) {
			items(
				offset: $offset
				first: $count
				orderBy: PUBLISH_DATE_DESC
				filter: { isPublished: { equalTo: true } }
			) {
				nodes {
					programSet {
						title
					}
					audios {
						title
						downloadUrl,
						allowDownload,
						url
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

func getCollectionUrls(queryId string) ([]File, error) {
	query := `query EpisodesQuery($id: ID!, $offset: Int!, $limit: Int!) {
		result: editorialCollection(id: $id, offset: $offset, limit: $limit) {
			items {
				nodes {
			  		id
					programSet {
						title
					}
			  		audios {
						title
						downloadUrl
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

	files := extractDownloadUrls(&response)

	return files, nil
}

func getEpisodeUrls(queryId string) ([]File, error) {
	type ItemResponse struct {
		Result struct {
			ProgramSet struct {
				Title string
			}
			Audios []struct {
				Title       string
				DownloadUrl *string
				Url         string
			}
		}
	}

	query := `query EpisodeQuery($id: ID!) {
		result: item(id: $id) {
			programSet {
				title
			}
		  	audios {
				title
				downloadUrl
				url
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

	var files []File
	for _, audios := range response.Result.Audios {
		if audios.DownloadUrl != nil {
			if *audios.DownloadUrl != "" {
				file := File{url: *audios.DownloadUrl, fileName: toFileName(audios.Title, response.Result.ProgramSet.Title)}
				files = append(files, file)
			} else {
				file := File{url: *&audios.Url, fileName: toFileName(audios.Title, response.Result.ProgramSet.Title)}
				files = append(files, file)
			}
		}
	}

	return files, nil
}

func sendGraphQlQuery(query string, variables map[string]interface{}, response interface{}) error {
	client := graphql.NewClient(graphqlEndPoint, nil)

	rawGraphqlResponse, graphQlErr := client.ExecRaw(context.Background(), query, variables)
	if graphQlErr != nil {
		return graphQlErr
	}

	if jsonError := json.Unmarshal(rawGraphqlResponse, &response); jsonError != nil {
		return jsonError
	}

	return nil
}

func getDownloadUrls(url string) ([]File, error) {
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

var DownloadCmd = &cobra.Command{
	Use:   "download [URL] [targetDirectory]",
	Short: "Download all episodes of a program/collection or an individual episode in the ARD Audiothek.",
	Long:  "Download all episodes of a program/collection or an individual episode in the ARD Audiothek. Limited to 100 episodes.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		run(args[0], args[1])
	},
}

func run(url string, targetDirectory string) {
	files, err := getDownloadUrls(url)

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		downloadErr := downloadFile(file, targetDirectory)

		if downloadErr != nil {
			fmt.Printf("Downloading file %s failed with error: %v\n", file.url, downloadErr)
		}
	}
}

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

func init() {
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:   "download [URL] [targetDirectory]",
	Short: "Download all episodes of a given program",
	Long:  "Download all episodes of a given program to the given target directory. Limited to 100 episodes.",
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

func extractDownloadUrls(response *Response) []string {
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

func extractProgramId(url string) string {
	normalizedUrl := strings.TrimSuffix(url, "/")
	pattern := regexp.MustCompile(`^https:\/\/www\.ardaudiothek\.de\/sendung\/.*\/(\d*)$`)
	return pattern.FindStringSubmatch(normalizedUrl)[1]
}

type Response struct {
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

func run(url string, targetDirectory string) {
	programId := extractProgramId(url)

	client := graphql.NewClient("https://api.ardaudiothek.de/graphql", nil)

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
		"id":     programId,
		"offset": 0,
		"count":  100,
	}

	rawGraphqlResponse, graphQlErr := client.ExecRaw(context.Background(), query, variables)
	if graphQlErr != nil {
		panic(graphQlErr)
	}

	var response Response
	if jsonError := json.Unmarshal(rawGraphqlResponse, &response); jsonError != nil {
		panic(jsonError)
	}

	urls := extractDownloadUrls(&response)

	for _, url := range urls {
		downloadErr := downloadFile(url, targetDirectory)

		if downloadErr != nil {
			fmt.Printf("Downloading URL %s failed with error: %v", url, downloadErr)
		}
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/meilisearch/meilisearch-go"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	// GitHub Authentication
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := githubv4.NewClient(tc)

	// Meilisearch Configuration
	ms := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: os.Getenv("MEILISEARCH_HOST"),
	})

	// GitHub username to query
	username := os.Getenv("TARGET_USERNAME")

	// Fetching organizations
	var orgQuery struct {
		User struct {
			Organizations struct {
				Nodes []struct {
					Login string
				}
			} `graphql:"organizations(first: 100)"`
		} `graphql:"user(login: $username)"`
	}
	orgVariables := map[string]interface{}{
		"username": githubv4.String(username),
	}
	var orgNames []string
	err := client.Query(ctx, &orgQuery, orgVariables)
	if err != nil {
		log.Fatal(err)
	}
	for _, org := range orgQuery.User.Organizations.Nodes {
		orgNames = append(orgNames, org.Login)
	}

	// Fetching repositories for each organization
	var query struct {
		Organization struct {
			Repositories struct {
				Nodes []struct {
					Name  string
					Owner struct {
						Login string
					}
					StargazerCount   int
					URL              string
					Description      string
					RepositoryTopics struct {
						Nodes []struct {
							Topic struct {
								Name string
							}
						}
					} `graphql:"repositoryTopics(first: 10)"`
					Releases struct {
						Nodes []struct {
							TagName string
						}
					} `graphql:"releases(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
				}
			} `graphql:"repositories(first: 100)"`
		} `graphql:"organization(login: $orgName)"`
	}
	var repositories []struct {
		Name           string
		Owner          string
		StargazerCount int
		URL            string
		Description    string
		Topics         []string
		TagName        string
	}
	for _, orgName := range orgNames {
		variables := map[string]interface{}{
			"orgName": githubv4.String(orgName),
		}
		err := client.Query(ctx, &query, variables)
		if err != nil {
			log.Fatal(err)
		}
		for _, node := range query.Organization.Repositories.Nodes {
			topics := make([]string, 0, len(node.RepositoryTopics.Nodes))
			for _, topic := range node.RepositoryTopics.Nodes {
				topics = append(topics, topic.Topic.Name)
			}
			tagName := ""
			if len(node.Releases.Nodes) > 0 {
				tagName = node.Releases.Nodes[0].TagName
			}
			repositories = append(repositories, struct {
				Name           string
				Owner          string
				StargazerCount int
				URL            string
				Description    string
				Topics         []string
				TagName        string
			}{
				Name:           node.Name,
				Owner:          node.Owner.Login,
				StargazerCount: node.StargazerCount,
				URL:            node.URL,
				Description:    node.Description,
				Topics:         topics,
				TagName:        tagName,
			})
		}
	}

	// Fetching starred repositories
	var starredQuery struct {
		User struct {
			StarredRepositories struct {
				Nodes []struct {
					Name  string
					Owner struct {
						Login string
					}
					StargazerCount   int
					URL              string
					Description      string
					RepositoryTopics struct {
						Nodes []struct {
							Topic struct {
								Name string
							}
						}
					} `graphql:"repositoryTopics(first: 10)"`
					Releases struct {
						Nodes []struct {
							TagName string
						}
					} `graphql:"releases(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
				}
			} `graphql:"starredRepositories(first: 100)"`
		} `graphql:"user(login: $username)"`
	}
	starredVariables := map[string]interface{}{
		"username": githubv4.String(username),
	}
	err = client.Query(ctx, &starredQuery, starredVariables)
	if err != nil {
		log.Fatal(err)
	}
	for _, node := range starredQuery.User.StarredRepositories.Nodes {
		topics := make([]string, 0, len(node.RepositoryTopics.Nodes))
		for _, topic := range node.RepositoryTopics.Nodes {
			topics = append(topics, topic.Topic.Name)
		}
		tagName := ""
		if len(node.Releases.Nodes) > 0 {
			tagName = node.Releases.Nodes[0].TagName
		}
		repositories = append(repositories, struct {
			Name           string
			Owner          string
			StargazerCount int
			URL            string
			Description    string
			Topics         []string
			TagName        string
		}{
			Name:           node.Name,
			Owner:          node.Owner.Login,
			StargazerCount: node.StargazerCount,
			URL:            node.URL,
			Description:    node.Description,
			Topics:         topics,
			TagName:        tagName,
		})
	}

	// Indexing data into Meilisearch
	indexName := "github_stars"
	index := ms.Index(indexName)

	// Prepare documents for indexing
	var documents []map[string]interface{}
	for i, repo := range repositories {
		document := map[string]interface{}{
			"id":              i + 1,
			"repository_name": string(repo.Name),
			"owner":           string(repo.Owner),
			"stars":           int(repo.StargazerCount),
			"url":             string(repo.URL),
			"description":     string(repo.Description),
			"topics":          []string(repo.Topics),
			"tag_name":        string(repo.TagName),
		}
		documents = append(documents, document)
	}

	task, err := index.AddDocuments(documents)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Data indexed successfully! JobID: " + strconv.FormatInt(task.TaskUID, 10) + ", documents: " + strconv.Itoa(len(documents)))
}

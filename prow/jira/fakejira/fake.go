/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fakejira

import (
	"context"
	"fmt"
	"k8s.io/klog"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/sirupsen/logrus"
	jiraclient "k8s.io/test-infra/prow/jira"
)

type FakeClient struct {
	ExistingIssues []jira.Issue
	ExistingLinks  map[string][]jira.RemoteLink
	NewLinks       []jira.RemoteLink
	GetIssueError  error
	IssueSearch    []jira.Issue
}

type JqlQueries struct {
	Jql []*JqlQuery
}

type JqlQuery struct {
	Key      string
	Operator string
	Value    string
}

var supportedJQLKeys = []string{"project", "category", "id", "issueType", "status"}

var jqlOperators = map[string]string{
	"equal":    "=",
	"contains": " IN ",
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func SplitJQLQuery(s string) (JqlQuery, error) {
	var min int
	var name string
	i := 0
	for operatorName, operatorType := range jqlOperators {
		ind := strings.Index(s, operatorType)
		if ind == -1 {
			continue
		}
		if i == 0 {
			min = ind
			name = operatorName
		}
		if ind < min && ind > -1 {
			min = ind
			name = operatorName
		}
		i++
	}
	if contains(supportedJQLKeys, s[:min]) {
		return JqlQuery{
			Key:      s[:min],
			Operator: name,
			Value:    trimQuotes(s[min+len(jqlOperators[name]):]),
		}, nil
	} else {
		return JqlQuery{}, fmt.Errorf("failed to split the jql string: %s", s)
	}

}

func jqlContains(s string, str string) bool {
	if s[0:1] != "(" && s[len(s)-1:len(s)] != ")" {
		klog.Warningf("incorrect format of the JQL query: %s", s)
		return false
	}
	a := strings.Split(s[1:len(s)-1], ",")
	for _, v := range a {
		if strings.TrimSpace(v) == strings.TrimSpace(str) {
			return true
		}
	}
	return false
}

// ParseJQL
// Supported JQL Operators: "&" (and); JQL fields Operators: "=" (equal), " IN " (contains); JQL Keys: supportedJQLKeys
func ParseJQL(jql string) JqlQueries {
	jqlParts := strings.Split(jql, "&")
	query := JqlQueries{}
	for _, part := range jqlParts {
		jql, err := SplitJQLQuery(part)
		if err == nil {
			query.Jql = append(query.Jql, &jql)
		} else {
			klog.Warning(err)
			continue
		}
	}
	return query
}

// Search
// Supported JQL Operators: "&" (and); JQL fields Operators: "=" (equal), " IN " (contains); JQL Keys: supportedJQLKeys
func (f *FakeClient) Search(jql string, options *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
	return f.SearchWithContext(context.Background(), jql, options)
}

// SearchWithContext
// Supported JQL Operators: "&" (and); JQL fields Operators: "=" (equal), " IN " (contains); JQL Keys: supportedJQLKeys
func (f *FakeClient) SearchWithContext(ctx context.Context, jql string, options *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
	ctx = context.Background()
	if f.GetIssueError != nil {
		return nil, nil, f.GetIssueError
	}
	var issueList []jira.Issue
	var tmpIssueList []jira.Issue
	filters := ParseJQL(jql)
	for _, filter := range filters.Jql {
		if len(issueList) == 0 {
			tmpIssueList = f.IssueSearch
		} else {
			tmpIssueList = issueList
			issueList = []jira.Issue{}
		}
		for _, existingIssue := range tmpIssueList {
			switch {
			case filter.Key == "id":
				if filter.Operator == "equal" {
					if existingIssue.ID == filter.Value || existingIssue.Key == filter.Value {
						issueList = append(issueList, existingIssue)
					}
				}
				if filter.Operator == "contains" {
					if jqlContains(filter.Value, existingIssue.ID) || jqlContains(filter.Value, existingIssue.Key) {
						issueList = append(issueList, existingIssue)
					}
				}
			case filter.Key == "project":
				if filter.Operator == "equal" {
					if existingIssue.Fields.Project.Name == filter.Value {
						issueList = append(issueList, existingIssue)
					}
				}
				if filter.Operator == "contains" {
					if jqlContains(filter.Value, existingIssue.Fields.Project.Name) {
						issueList = append(issueList, existingIssue)
					}
				}
			case filter.Key == "category":
				if filter.Operator == "equal" {
					if existingIssue.Fields.Project.ProjectCategory.Name == filter.Value {
						issueList = append(issueList, existingIssue)
					}
				}
				if filter.Operator == "contains" {
					if jqlContains(filter.Value, existingIssue.Fields.Project.ProjectCategory.Name) {
						issueList = append(issueList, existingIssue)
					}
				}
			case filter.Key == "issueType":
				if filter.Operator == "equal" {
					if existingIssue.Fields.Type.Name == filter.Value {
						issueList = append(issueList, existingIssue)
					}
				}
				if filter.Operator == "contains" {
					if jqlContains(filter.Value, existingIssue.Fields.Type.Name) {
						issueList = append(issueList, existingIssue)
					}
				}
			case filter.Key == "status":
				if filter.Operator == "equal" {
					if existingIssue.Fields.Status.Name == filter.Value {
						issueList = append(issueList, existingIssue)
					}
				}
				if filter.Operator == "contains" {
					if jqlContains(filter.Value, existingIssue.Fields.Status.Name) {
						issueList = append(issueList, existingIssue)
					}
				}
			}
		}
	}
	if len(issueList) == 0 {
		return nil, nil, jiraclient.NewNotFoundError(fmt.Errorf("no issue found with the jql query: %s", jql))
	}
	if options.MaxResults == 0 {
		options.MaxResults = 50
	}
	limit := options.MaxResults
	if len(issueList) < options.MaxResults {
		limit = len(issueList)
	}
	return issueList[0:limit], &jira.Response{MaxResults: options.MaxResults, StartAt: options.StartAt, Total: len(issueList)}, nil

}

func (f *FakeClient) ListProjects() (*jira.ProjectList, error) {
	return nil, nil
}

func (f *FakeClient) GetIssue(id string) (*jira.Issue, error) {
	if f.GetIssueError != nil {
		return nil, f.GetIssueError
	}
	for _, existingIssue := range f.ExistingIssues {
		if existingIssue.ID == id {
			return &existingIssue, nil
		}
	}
	return nil, jiraclient.NewNotFoundError(fmt.Errorf("No issue %s found", id))
}

func (f *FakeClient) GetRemoteLinks(id string) ([]jira.RemoteLink, error) {
	return f.ExistingLinks[id], nil
}

func (f *FakeClient) AddRemoteLink(id string, link *jira.RemoteLink) error {
	if _, err := f.GetIssue(id); err != nil {
		return err
	}
	f.NewLinks = append(f.NewLinks, *link)
	return nil
}

func (f *FakeClient) JiraClient() *jira.Client {
	panic("not implemented")
}

const FakeJiraUrl = "https://my-jira.com"

func (f *FakeClient) JiraURL() string {
	return FakeJiraUrl
}

func (f *FakeClient) UpdateRemoteLink(id string, link *jira.RemoteLink) error {
	if _, err := f.GetIssue(id); err != nil {
		return err
	}
	if _, found := f.ExistingLinks[id]; !found {
		return jiraclient.NewNotFoundError(fmt.Errorf("Link for issue %s not found", id))
	}
	f.NewLinks = append(f.NewLinks, *link)
	return nil
}

func (f *FakeClient) Used() bool {
	return true
}

func (f *FakeClient) WithFields(fields logrus.Fields) jiraclient.Client {
	return f
}

func (f *FakeClient) ForPlugin(string) jiraclient.Client {
	return f
}

package fakejira

import (
	"context"
	"github.com/andygrunwald/go-jira"
	"reflect"
	"testing"
)

func TestFakeClient_SearchWithContext(t *testing.T) {
	s := make(map[Request]Response)
	issueList := []jira.Issue{
		{
			ID:     "123",
			Fields: &jira.IssueFields{Project: jira.Project{Name: "test"}},
		},
		{
			ID:     "1234",
			Fields: &jira.IssueFields{Project: jira.Project{Name: "test"}},
		},
		{
			ID:     "12345",
			Fields: &jira.IssueFields{Project: jira.Project{Name: "test"}},
		},
	}
	searchOptions := &jira.SearchOptions{MaxResults: 50, StartAt: 0}
	request := Request{
		query:   "project=test",
		options: searchOptions,
	}
	response := Response{
		values:   issueList,
		response: &jira.Response{StartAt: 0, MaxResults: 3, Total: 3},
		error:    nil,
	}
	s[request] = response
	fakeClient := &FakeClient{Responses: s}
	r, v, err := fakeClient.SearchWithContext(context.Background(), "project=test", searchOptions)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if !reflect.DeepEqual(r, issueList) || !reflect.DeepEqual(&jira.Response{StartAt: 0, MaxResults: 3, Total: 3}, v) {
		t.Fatalf("unexpected response")
	}
	r, v, err = fakeClient.SearchWithContext(context.Background(), "unknown_query=test", searchOptions)
	if r != nil && v != nil && err == nil {
		t.Fatal("unexpected result")
	}
}

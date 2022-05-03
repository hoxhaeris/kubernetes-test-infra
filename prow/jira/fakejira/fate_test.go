package fakejira

import (
	"github.com/andygrunwald/go-jira"
	"testing"
)

func TestClient_ParseJQL(t *testing.T) {
	validQueries := []struct {
		query string
		jql   []JqlQuery
	}{
		{
			query: "project='Openshift \\=Container=Platform'",
			jql: []JqlQuery{
				{
					Key:      "project",
					Operator: "equal",
					Value:    "'Openshift \\=Container=Platform'",
				},
			},
		},
		{
			query: "id IN (123,321,OCP-IN12)",
			jql: []JqlQuery{
				{
					Key:      "id",
					Operator: "contains",
					Value:    "(123,321,OCP-IN12)",
				},
			},
		},
		{
			query: "project='Openshift Continuous Release'&id IN (123,321,124)",
			jql: []JqlQuery{
				{
					Key:      "project",
					Operator: "equal",
					Value:    "'Openshift \\=Container=Platform'",
				},
				{
					Key:      "id",
					Operator: "contains",
					Value:    "(123,321,OCP-IN12)",
				},
			},
		},
	}
	for _, tc := range validQueries {
		t.Run(tc.query, func(t *testing.T) {
			a := ParseJQL(tc.query)
			if len(a.Jql) != len(tc.jql) {
				t.Fatal("unexpected length of the JQL array")
			}
			for i, query := range tc.jql {
				if a.Jql[i].Key != query.Key || query.Operator != query.Operator || query.Value != query.Value {
					t.Fatalf("unexpected value for the key/value/operator field. %v", a.Jql[0])
				}

			}
		})
	}
}

func TestClient_Search(t *testing.T) {
	issueList := []jira.Issue{
		{
			ID:     "123",
			Fields: &jira.IssueFields{Project: jira.Project{Name: "Test1"}},
		},
		{
			ID:     "1234",
			Fields: &jira.IssueFields{Project: jira.Project{Name: "Test2"}},
		},
		{
			ID:     "12345",
			Fields: &jira.IssueFields{Project: jira.Project{Name: "Test3"}},
		},
	}

	testCases := []struct {
		jql             string
		maxResult       int
		expectedResults int
		totalResults    int
	}{
		{
			jql:             "id IN (123,1234)&project='Test1'",
			maxResult:       0,
			expectedResults: 1,
			totalResults:    1,
		},
		{
			jql:             "id IN (123,1234)",
			maxResult:       1,
			expectedResults: 1,
			totalResults:    2,
		},
	}

	fakeClient := &FakeClient{IssueSearch: issueList}
	for _, tc := range testCases {
		t.Run(tc.jql, func(t *testing.T) {
			search, r, err := fakeClient.Search(tc.jql, &jira.SearchOptions{MaxResults: tc.maxResult})
			if tc.maxResult == 0 {
				tc.maxResult = 50
			}
			if len(search) != 1 && r.MaxResults != tc.maxResult {
				t.Fatal("unexpected number of issues returned")
			}
			if r.Total != tc.totalResults {
				t.Fatal("unexpected number of total found issues")
			}
			if err != nil {
				t.Fatal("unexpected error returned")
			}
		})
	}
}

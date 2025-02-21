package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// CV represents the overall structure of the Comic Vine API response.
type CV struct {
	Error                string     `json:"error"`
	Limit                int        `json:"limit"`
	Offset               int        `json:"offset"`
	NumberOfPageResults  int        `json:"number_of_page_results"`
	NumberOfTotalResults int        `json:"number_of_total_results"`
	StatusCode           int        `json:"status_code"`
	Results              []CVResult `json:"results"` // Use the named struct
}

// CVResult represents a single issue result from the Comic Vine API.
type CVResult struct {
	APIDetailURL  string    `json:"api_detail_url"`
	ID            int       `json:"id"`
	IssueNumber   string    `json:"issue_number"`
	Name          string    `json:"name"` //  Use string, handle potential null in display logic
	SiteDetailURL string    `json:"site_detail_url"`
	Description   string    `json:"description"`
	Credits       []Credits `json:"person_credits"`
	StoreDate     string    `json:"store_date"`
	Volume        struct {
		APIDetailURL  string `json:"api_detail_url"`
		ID            int    `json:"id"`
		Name          string `json:"name"`
		SiteDetailURL string `json:"site_detail_url"` // Added, consistent with other structs
	} `json:"volume"`
	ResourceType string `json:"resource_type"` // If you need this, keep it.
}

type Credits struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

const (
	cvURL   = "https://comicvine.gamespot.com/api/"      // Use HTTPS
	api_key = "" // Replace with your API key
)

func cvGetCredits(id int) ([]Credits, error) {
	urlCV := fmt.Sprintf("%sissue/4000-%v?api_key=%s&format=json&field_list=person_credits", cvURL, id, api_key)
	resp, err := http.Get(urlCV)
	if err != nil {
		return []Credits{}, err
	}
	v := struct {
		Error      string `json:"error"`
		StatusCode int    `json:"status_code"`
		Results    struct {
			Credits []Credits `json:"person_credits"`
		} `json:"results"` // Use the named struct
	}{}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&v)
	return v.Results.Credits, nil
}

func cvSearch(issueName string, issueNumber string) (*CVResult, error) {
	query := issueName + " " + issueNumber
	urlCV := fmt.Sprintf(cvURL + "search")
	req, err := http.NewRequest(http.MethodGet, urlCV, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	q := req.URL.Query()
	q.Add("resources", "issue") // Comic Vine uses "resource_type", not "resources"
	q.Add("format", "json")
	q.Add("api_key", api_key)
	q.Add("limit", "20")
	q.Add("field_list", "id,name,issue_number,volume,api_detail_url,site_detail_url,description,person_credits") // Corrected field list
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close() // Close the response body

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var cv CV
	err = json.NewDecoder(resp.Body).Decode(&cv)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	if cv.NumberOfTotalResults == 0 {
		return nil, fmt.Errorf("no results found for query: %s", query)
	}

	if cv.Error != "OK" {
		return nil, fmt.Errorf("API returned an error: %s", cv.Error)
	}

	// Display results and ask the user to choose
	fmt.Printf("Issue name entered: %v", issueName+" -- "+"Issue number: #"+issueNumber+"\n")
	fmt.Println("Search Results:")
	for i, result := range cv.Results {
		// Handle potentially missing issue names gracefully.
		issueName := result.Name
		if issueName == "" {
			issueName = "(No issue title available)"
		}
		volumeName := result.Volume.Name
		if volumeName == "" {
			volumeName = "(No volume name available)"
		}
		fmt.Printf("%d: %s -- Issue Number: #%s -- %s\n", i+1, volumeName, result.IssueNumber, issueName)
		fmt.Printf("    Site detail URL: %s\n", result.SiteDetailURL)
		fmt.Printf("    Volume Name: %s\n", result.Volume.Name)
		fmt.Printf("       Volume detail URL: %s\n", result.Volume.SiteDetailURL)

	}

	var choiceStr string // Use a string to handle non-numeric input
	var choice int
	for {
		fmt.Printf("Enter the number of the correct issue (or 0 to cancel): ")
		_, err := fmt.Scanln(&choiceStr) // Read user input as a string
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}

		// Try to convert the input to an integer
		choice, err = strconv.Atoi(choiceStr)
		if err != nil || choice < 0 || choice > len(cv.Results) {
			fmt.Println("Invalid input. Please enter a valid number.")
			continue
		}
		break
	}

	if choice == 0 {
		return nil, fmt.Errorf("operation canceled by user")
	}

	return &cv.Results[choice-1], nil
}

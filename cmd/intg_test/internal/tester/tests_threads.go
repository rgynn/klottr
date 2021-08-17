package tester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rgynn/klottr/pkg/thread"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
)

func (tester *Tester) createThread(token *string, category string) (*thread.Model, error) {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s", tester.cfg.Addr, category)

	reqbody, err := json.Marshal(&thread.Model{
		Username: ptrconv.StringPtr(tester.username),
		Title:    ptrconv.StringPtr("test title"),
		URL:      ptrconv.StringPtr("https://klottr.com"),
		Content:  `fawfwefawlfuwaelfuhwaelfuhwaelfiuhwalf`,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqbody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := tester.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		break
	default:
		return nil, fmt.Errorf("expected status %d in deactivate response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	var result *thread.Model
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	tester.logger.Infof("OK: Thread created, title: %s, slug_id: %s, slug_title: %s\n", "test title", *result.SlugID, *result.SlugTitle)

	return result, nil
}

func (tester *Tester) listThreads(token *string, category string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s", tester.cfg.Addr, category)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := tester.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		break
	default:
		return fmt.Errorf("expected status %d in list threads response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	var response []*thread.Model
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	tester.logger.Infof("OK: List threads in category: %s", category)

	return nil
}

func (tester *Tester) getThread(token *string, category string, slugID, slugTitle *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s/t/%s/%s", tester.cfg.Addr, category, *slugID, *slugTitle)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := tester.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		break
	default:
		return fmt.Errorf("expected status %d in get thread response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	var response *thread.Model
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	if *response.SlugID != *slugID {
		return fmt.Errorf("expected slug_id to be: %s, got: %s", *slugID, *response.SlugID)
	}

	if *response.SlugTitle != *slugTitle {
		return fmt.Errorf("expected slug_title to be: %s, got: %s", *slugTitle, *response.SlugTitle)
	}

	tester.logger.Infof("OK: Got thread, category: %s, slug_id: %s, slug_title: %s\n", category, *response.SlugID, *response.SlugTitle)

	return nil
}

func (tester *Tester) upvoteThread(token *string, category string, slugID, slugTitle *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s/t/%s/%s/vote", tester.cfg.Addr, category, *slugID, *slugTitle)

	reqbody, err := json.Marshal(&user.Vote{
		SlugType: ptrconv.StringPtr("threads"),
		SlugID:   slugID,
		Value:    ptrconv.Int8Ptr(1),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqbody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := tester.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusAccepted:
		break
	default:
		return fmt.Errorf("expected status %d in upvote response, got: %d, body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	tester.logger.Infof("OK: Thread upvoted")

	return nil
}

func (tester *Tester) downvoteThread(token *string, category string, slugID, slugTitle *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s/t/%s/%s/vote", tester.cfg.Addr, category, *slugID, *slugTitle)

	reqbody, err := json.Marshal(&user.Vote{
		SlugType: ptrconv.StringPtr("threads"),
		SlugID:   slugID,
		Value:    ptrconv.Int8Ptr(-1),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqbody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := tester.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusAccepted:
		break
	default:
		return fmt.Errorf("expected status %d in upvote response, got: %d, body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	tester.logger.Infof("OK: Thread downvoted")

	return nil
}

func (tester *Tester) validateVotes(token *string, category string, slugID, slugTitle *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s/t/%s/%s", tester.cfg.Addr, category, *slugID, *slugTitle)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := tester.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		break
	default:
		return fmt.Errorf("expected status %d in get thread response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	var response *thread.Model
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	if response.Counters.Votes != 0 {
		return fmt.Errorf("expected num votes to be 0, got: %d, body: %s", response.Counters.Votes, string(body))
	}

	tester.logger.Infof("OK: Num votes validated")

	return nil
}

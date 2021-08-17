package tester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rgynn/klottr/pkg/comment"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
)

func (tester *Tester) createComment(token *string, category string, slugID, slugTitle *string) (*comment.Model, error) {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s/t/%s/%s/comments", tester.cfg.Addr, category, *slugID, *slugTitle)

	reqbody, err := json.Marshal(&comment.Model{
		Content: `test comment`,
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
		return nil, fmt.Errorf("expected status %d in create comment response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	var result *comment.Model
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	tester.logger.Infof("OK: Comment created")

	return result, nil
}

func (tester *Tester) upvoteComment(token *string, category string, slugID, slugTitle, cmntSlugID *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s/t/%s/%s/comments/%s/vote", tester.cfg.Addr, category, *slugID, *slugTitle, *cmntSlugID)

	reqbody, err := json.Marshal(&user.Vote{
		SlugType: ptrconv.StringPtr("comments"),
		SlugID:   cmntSlugID,
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
		return fmt.Errorf("expected status %d in comment upvote response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	tester.logger.Infof("OK: Comment upvoted")

	return nil
}

func (tester *Tester) downvoteComment(token *string, category string, slugID, slugTitle, cmntSlugID *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s/t/%s/%s/comments/%s/vote", tester.cfg.Addr, category, *slugID, *slugTitle, *cmntSlugID)

	reqbody, err := json.Marshal(&user.Vote{
		SlugType: ptrconv.StringPtr("comments"),
		SlugID:   cmntSlugID,
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
		return fmt.Errorf("expected status %d in comment downvote response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	tester.logger.Infof("OK: Comment downvoted")

	return nil
}

func (tester *Tester) validateCommentVotes(token *string, category string, slugID, slugTitle, cmntSlugID *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s/t/%s/%s/comments/%s", tester.cfg.Addr, category, *slugID, *slugTitle, *cmntSlugID)

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
		return fmt.Errorf("expected status %d in get comment response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	var response *comment.Model
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	tester.logger.Infof("OK: Comment votes validated")

	return nil
}

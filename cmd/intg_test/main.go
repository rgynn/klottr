package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/rgynn/klottr/pkg/api"
	"github.com/rgynn/klottr/pkg/config"
	"github.com/rgynn/klottr/pkg/thread"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
	"github.com/sirupsen/logrus"
)

var (
	logger     = logrus.New()
	cfg        *config.Config
	client     *http.Client
	err        error
	categories = []string{"misc"}
)

const (
	username = "testuser"
	password = "testpasswd"
)

func main() {

	client = &http.Client{
		Timeout: time.Second * 5,
	}

	cfg, err = config.NewFromEnv()
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.Infof("Integration test started")

	if err := signupTestUser(); err != nil {
		logger.Fatal(err)
	}

	token, err := signinTestUser(http.StatusOK)
	if err != nil {
		logger.Fatal(err)
	}

	if err := validateJWT(token); err != nil {
		logger.Fatal(err)
	}

	for _, category := range categories {
		thrd, err := createThread(token, category)
		if err != nil {
			logger.Fatal(err)
		}
		if err := getThread(token, category, thrd.SlugID, thrd.SlugTitle); err != nil {
			logger.Fatal(err)
		}
		if err := listThreads(token, category); err != nil {
			logger.Fatal(err)
		}
	}

	if err := deactivateTestUser(token); err != nil {
		logger.Fatal(err)
	}

	if _, err = signinTestUser(http.StatusUnauthorized); err != nil {
		logger.Fatal(err)
	}
}

func signupTestUser() error {

	url := fmt.Sprintf("http://%s/auth/signup", cfg.Addr)

	reqbody, err := json.Marshal(&user.Model{
		Username: ptrconv.StringPtr(username),
		Password: ptrconv.StringPtr(password),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqbody))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		logger.Infof("OK: Signed up username: %s\n", username)
		return nil
	case http.StatusConflict:
		logger.Warnf("WARN: Conflict, user already present: %s\n", username)
		return nil
	default:
		return fmt.Errorf("expected status %d in signup response, got: %d, response body: %s", http.StatusCreated, resp.StatusCode, string(body))
	}
}

func signinTestUser(expectedStatusCode int) (*string, error) {

	url := fmt.Sprintf("http://%s/auth/signin", cfg.Addr)

	reqbody, err := json.Marshal(&api.LoginInput{
		Username: ptrconv.StringPtr(username),
		Password: ptrconv.StringPtr(password),
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqbody))
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("expected status %d in signin response, got: %d, response body: %s", http.StatusOK, resp.StatusCode, string(body))
	}

	if expectedStatusCode == http.StatusUnauthorized {
		logger.Infof("OK: Could not sign in deactivated user: %s", username)
		return nil, nil
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	token, ok := response["token"].(string)
	if !ok {
		return nil, fmt.Errorf("expected response to contain a token string")
	}

	logger.Infof("OK: Signed in username: %s", username)

	return &token, nil
}

func validateJWT(token *string) error {

	claims := new(api.JWTClaims)

	_, err := jwt.ParseWithClaims(*token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return err
	}

	logger.Infof("OK: JWT valid")

	return nil
}

func deactivateTestUser(token *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/users/%s/deactivate", cfg.Addr, username)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := client.Do(req)
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
		return fmt.Errorf("expected status %d in deactivate response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	logger.Infof("OK: Deactivated username: %s\n", username)

	return nil
}

func createThread(token *string, category string) (*thread.Model, error) {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s", cfg.Addr, category)

	reqbody, err := json.Marshal(&thread.Model{
		Username: ptrconv.StringPtr(username),
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

	resp, err := client.Do(req)
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

	logger.Infof("OK: Thread created, title: %s, slug_id: %s, slug_title: %s\n", "test title", *result.SlugID, *result.SlugTitle)

	return result, nil
}

func listThreads(token *string, category string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s", cfg.Addr, category)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := client.Do(req)
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

	logger.Infof("OK: List threads in category: %s", category)

	return nil
}

func getThread(token *string, category string, slugID, slugTitle *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/c/%s/t/%s/%s", cfg.Addr, category, *slugID, *slugTitle)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := client.Do(req)
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

	if *response.Category != category {
		return fmt.Errorf("expected category to be: %s, got: %s", category, *response.Category)
	}

	if *response.SlugID != *slugID {
		return fmt.Errorf("expected slug_id to be: %s, got: %s", *slugID, *response.SlugID)
	}

	if *response.SlugTitle != *slugTitle {
		return fmt.Errorf("expected slug_title to be: %s, got: %s", *slugTitle, *response.SlugTitle)
	}

	logger.Infof("OK: Got thread, category: %s, slug_id: %s, slug_title: %s\n", category, *response.SlugID, *response.SlugTitle)

	return nil
}
